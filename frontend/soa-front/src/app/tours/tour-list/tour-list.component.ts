import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { TourService, Tour } from '../../services/tour.service';
import { AuthService } from '../../auth/services/auth.service';

@Component({
  selector: 'app-tour-list',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './tour-list.component.html',
  styleUrl: './tour-list.component.css'
})
export class TourListComponent implements OnInit {
  tours: Tour[] = [];
  loading = false;
  error = '';
  currentUser: any;

  constructor(private tourService: TourService, private authService: AuthService) {}

  ngOnInit(): void {
    this.authService.currentUser$.subscribe(user => {
      this.currentUser = user;
      if (user) this.loadTours(user.username);
    });
  }

  loadTours(authorId: string): void {
    this.loading = true;
    this.tourService.getToursByAuthor(authorId).subscribe({
      next: (tours) => { this.tours = tours; this.loading = false; },
      error: () => { this.error = 'Greška pri učitavanju tura.'; this.loading = false; }
    });
  }

  difficultyLabel(d: string): string {
    return { easy: 'Lako', medium: 'Srednje', hard: 'Teško' }[d] ?? d;
  }
}