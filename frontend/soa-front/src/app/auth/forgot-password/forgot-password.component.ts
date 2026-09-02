import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { AuthService } from '../services/auth.service';

@Component({
  selector: 'app-forgot-password',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './forgot-password.component.html',
  styleUrl: '../auth-form.css'
})
export class ForgotPasswordComponent {
  form: FormGroup;
  loading = false;
  submitted = false;
  error = '';
  message = '';
  // Demo-only: there's no email service in this stack, so the backend
  // returns the reset token/link directly instead of emailing it.
  resetToken = '';

  constructor(private fb: FormBuilder, private authService: AuthService) {
    this.form = this.fb.group({
      usernameOrEmail: ['', [Validators.required]],
    });
  }

  get f() {
    return this.form.controls;
  }

  onSubmit(): void {
    this.submitted = true;
    this.error = '';
    this.message = '';
    this.resetToken = '';

    if (this.form.invalid) {
      return;
    }

    this.loading = true;
    this.authService.requestPasswordReset(this.form.value.usernameOrEmail).subscribe({
      next: (res) => {
        this.loading = false;
        this.message = res.message;
        this.resetToken = res.resetToken || '';
      },
      error: (err) => {
        this.loading = false;
        this.error = err.error?.error || 'Greška pri slanju zahteva.';
      }
    });
  }
}
