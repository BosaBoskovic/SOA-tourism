import { Component, OnInit } from '@angular/core';
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

  constructor(private profileService: ProfileService, private fb: FormBuilder) {}

  ngOnInit(): void {
    this.loading = true;
    this.profileService.getProfile().subscribe({
      next: (res) => {
        this.profile = res.profile;
        this.profileForm = this.fb.group({
          firstName: [res.profile.firstName],
          lastName: [res.profile.lastName],
          imageURL: [res.profile.imageURL],
          bio: [res.profile.bio],
          motto: [res.profile.motto],
        });
        this.loading = false;
      },
      error: () => {
        this.error = 'Greška pri učitavanju profila.';
        this.loading = false;
      }
    });
  }

  toggleEdit(): void {
    this.editMode = !this.editMode;
    this.success = '';
    this.error = '';
  }

  onSave(): void {
    this.saving = true;
    this.profileService.updateProfile(this.profileForm.value).subscribe({
      next: (res) => {
        this.profile = res.profile;
        this.editMode = false;
        this.saving = false;
        this.success = 'Profil uspješno ažuriran!';
      },
      error: () => {
        this.error = 'Greška pri ažuriranju profila.';
        this.saving = false;
      }
    });
  }
}