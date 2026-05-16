import { Component, EventEmitter, Input, Output, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

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
    private cdr: ChangeDetectorRef
  ) {}

  rating = 5;
  comment = '';
  visitDate = '';
  images: string[] = [];

  onImagesSelected(event: Event): void {
  const input = event.target as HTMLInputElement;

  if (!input.files || input.files.length === 0) {
    return;
  }

  Array.from(input.files).forEach(file => {
    const reader = new FileReader();

    reader.onload = () => {
      this.images = [...this.images, reader.result as string];
      this.cdr.detectChanges();
    };

    reader.readAsDataURL(file);
  });

  input.value = '';
}

  removeImage(index: number): void {
    this.images.splice(index, 1);
  }

  submit(): void {
    if (!this.comment.trim() || !this.visitDate) {
      alert('Popuni komentar i datum posete.');
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