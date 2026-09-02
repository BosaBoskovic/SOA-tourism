import { Component, EventEmitter, Input, Output, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { forkJoin } from 'rxjs';
import { ToastService } from '../../shared/toast/toast.service';
import { UploadService } from '../../services/upload.service';

export interface ReviewRequest {
  rating: number;
  comment: string;
  visitDate: string;
  images: string[];
}


@Component({
  selector: 'app-review-form',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './review-form.component.html',
  styleUrl: './review-form.component.css'
})
export class ReviewFormComponent {
  @Input() tour: any;
  @Output() close = new EventEmitter<void>();
  @Output() submitReview = new EventEmitter<ReviewRequest>();

  constructor(
    private cdr: ChangeDetectorRef,
    private toastService: ToastService,
    private uploadService: UploadService
  ) {}

  rating = 5;
  comment = '';
  visitDate = '';
  images: string[] = [];
  uploadingImages = false;

  onImagesSelected(event: Event): void {
  const input = event.target as HTMLInputElement;

  if (!input.files || input.files.length === 0) {
    return;
  }

  const files = Array.from(input.files);
  this.uploadingImages = true;

  forkJoin(files.map(file => this.uploadService.upload(file))).subscribe({
    next: (urls) => {
      this.images = [...this.images, ...urls];
      this.uploadingImages = false;
      this.cdr.detectChanges();
    },
    error: () => {
      this.uploadingImages = false;
      this.toastService.error('Greška pri otpremanju slika.');
      this.cdr.detectChanges();
    }
  });

  input.value = '';
}

  removeImage(index: number): void {
    this.images.splice(index, 1);
  }

  submit(): void {
    if (!this.comment.trim() || !this.visitDate) {
      this.toastService.error('Popuni komentar i datum posete.');
      return;
    }
    if (this.uploadingImages) {
      this.toastService.error('Sačekaj da se slike otpreme.');
      return;
    }

    this.submitReview.emit({
      rating: this.rating,
      comment: this.comment,
      visitDate: this.visitDate,
      images: this.images
    });
  }

  closeForm(): void {
    this.close.emit();
  }
}