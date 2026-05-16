import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { TourService } from '../../services/tour.service';
import { AuthService } from '../../auth/services/auth.service';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-tour-create',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink, FormsModule],
  templateUrl: './tour-create.component.html',
  styleUrl: './tour-create.component.css'
})
export class TourCreateComponent {
  form: FormGroup;
  loading = false;
  error = '';
  tagsInput = '';

  constructor(
    private fb: FormBuilder,
    private tourService: TourService,
    private authService: AuthService,
    private router: Router
  ) {
    this.form = this.fb.group({
      name: ['', [Validators.required]],
      description: ['', [Validators.required]],
      difficulty: ['', [Validators.required]],
      tags: [[]]
    });
  }

  addTag(): void {
    const tag = this.tagsInput.trim();
    if (tag) {
      const current: string[] = this.form.value.tags || [];
      this.form.patchValue({ tags: [...current, tag] });
      this.tagsInput = '';
    }
  }

  removeTag(tag: string): void {
    const current: string[] = this.form.value.tags || [];
    this.form.patchValue({ tags: current.filter(t => t !== tag) });
  }

  onSubmit(): void {
    if (this.form.invalid) return;
    this.loading = true;

    const user = this.authService['currentUserSubject'].getValue();
    this.tourService.createTour({ ...this.form.value, authorId: user.username }).subscribe({
      next: (tour) => this.router.navigate(['/tours', tour.id]),
      error: () => { this.error = 'Greška pri kreiranju ture.'; this.loading = false; }
    });
  }
}