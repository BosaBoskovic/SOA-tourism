import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { ProfileService, ProfileResponse } from '../services/profile.service';
import { UploadService } from '../services/upload.service';
import { ToastService } from '../shared/toast/toast.service';
import { SpinnerComponent } from '../shared/ui/spinner/spinner.component';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink, SpinnerComponent],
  templateUrl: './profile.component.html',
  styleUrl: './profile.component.css'
})
export class ProfileComponent implements OnInit {
  profile: ProfileResponse | null = null;
  profileForm!: FormGroup;
  editMode = false;
  loading = false;
  saving = false;
  uploadingImage = false;
  error = '';
  success = '';

  constructor(
    private profileService: ProfileService,
    private uploadService: UploadService,
    private toastService: ToastService,
    private fb: FormBuilder,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit(): void {
    this.loading = true;
    this.profileService.getProfile().subscribe({
      next: (res) => {
        this.profile = res.profile;
        this.profileForm = this.fb.group({
          firstName: [res.profile.firstName, [Validators.required, Validators.maxLength(80)]],
          lastName: [res.profile.lastName, [Validators.required, Validators.maxLength(80)]],
          imageURL: [res.profile.imageURL],
          bio: [res.profile.bio, [Validators.maxLength(500)]],
          motto: [res.profile.motto, [Validators.maxLength(200)]],
        });
        this.loading = false;
        this.cdr.detectChanges();
      },
      error: () => {
        this.error = 'Greška pri učitavanju profila.';
        this.loading = false;
        this.cdr.detectChanges();
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
    this.uploadingImage = true;
    this.uploadService.upload(file).subscribe({
      next: (url) => {
        this.profileForm.patchValue({ imageURL: url });
        this.uploadingImage = false;
        this.cdr.detectChanges();
      },
      error: () => {
        this.uploadingImage = false;
        this.toastService.error('Greška pri otpremanju slike.');
        this.cdr.detectChanges();
      }
    });
  }

  onSave(): void {
    if (this.profileForm.invalid) {
      this.profileForm.markAllAsTouched();
      return;
    }

    this.saving = true;
    this.error = '';
    this.profileService.updateProfile(this.profileForm.value).subscribe({
      next: (res) => {
        this.profile = res.profile;
        this.editMode = false;
        this.saving = false;
        this.success = 'Profil uspješno ažuriran!';
        this.cdr.detectChanges();
      },
      error: () => {
        this.error = 'Greška pri ažuriranju profila.';
        this.saving = false;
        this.cdr.detectChanges();
      }
    });
  }
}
