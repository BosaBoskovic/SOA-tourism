import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

// Small reusable "nothing here" state, extracted from blog's polished
// .empty-state styling - used wherever a list page has zero items instead
// of every page rolling its own empty-state markup/CSS.
@Component({
  selector: 'app-empty-state',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="app-empty-state">
      <span class="app-empty-icon" *ngIf="icon">{{ icon }}</span>
      <h3 *ngIf="title">{{ title }}</h3>
      <p *ngIf="message">{{ message }}</p>
    </div>
  `,
  styleUrl: './empty-state.component.css'
})
export class EmptyStateComponent {
  @Input() icon = '📭';
  @Input() title = '';
  @Input() message = '';
}
