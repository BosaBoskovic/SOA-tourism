import { Component, OnInit, ChangeDetectorRef, NgZone } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { ProfileService, ProfileResponse } from '../services/profile.service';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './profile.component.html',
  styleUrl: './profile.component.css'
})
export class ProfileComponent implements OnInit {
  profile: ProfileResponse | null = null;
  profileForm!: FormGroup;
  editMode = false;
  loading = false;
  saving = false;
  error = '';
  success = '';

  constructor(
  private profileService: ProfileService,
  private fb: FormBuilder,
  private cdr: ChangeDetectorRef,
  private zone: NgZone
) {}

  ngOnInit(): void {
    this.loading = true;
    this.profileService.getProfile().subscribe({
      next: (res) => {
  this.zone.run(() => {
    this.profile = res.profile;
    this.profileForm = this.fb.group({
      firstName: [res.profile.firstName],
      lastName: [res.profile.lastName],
      imageURL: [res.profile.imageURL],
      bio: [res.profile.bio],
      motto: [res.profile.motto],
    });
    this.loading = false;
    this.cdr.detectChanges();
  });
},
      error: () => {
  this.zone.run(() => {
    this.error = 'Greška pri učitavanju profila.';
    this.loading = false;
    this.cdr.detectChanges();
  });
}
    });
  }

  toggleEdit(): void {
    this.editMode = !this.editMode;
    this.success = '';
    this.error = '';
  }

  onProfileImageSelected(event: Event): void {
  const input = event.target as HTMLInputElement;

  if (!input.files || input.files.length === 0) {
    return;
  }

  const file = input.files[0];
  const reader = new FileReader();

  reader.onload = () => {
    this.profileForm.patchValue({
      imageURL: reader.result as string
    });

    this.cdr.detectChanges();
  };

  reader.readAsDataURL(file);
}

  onSave(): void {
    this.saving = true;
    this.profileService.updateProfile(this.profileForm.value).subscribe({
      next: (res) => {
  this.zone.run(() => {
    this.profile = res.profile;
    this.editMode = false;
    this.saving = false;
    this.success = 'Profil uspješno ažuriran!';
    this.cdr.detectChanges();
  });
},
      error: () => {
        this.error = 'Greška pri ažuriranju profila.';
        this.saving = false;
      }
    });
  }
}