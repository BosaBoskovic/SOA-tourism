import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

// Small reusable loading spinner, extracted from blog's polished
// .loading-state styling - used wherever a page shows an "Učitavanje..."
// state instead of every page rolling its own <p>Učitavanje...</p>.
@Component({
  selector: 'app-spinner',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="app-spinner-wrap">
      <div class="app-spinner"></div>
      <p *ngIf="label">{{ label }}</p>
    </div>
  `,
  styleUrl: './spinner.component.css'
})
export class SpinnerComponent {
  @Input() label = 'Učitavanje...';
}
