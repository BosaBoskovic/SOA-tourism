import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Review {
  id: string;
  tourId: string;
  touristId: string;
  touristName: string;
  rating: number;
  comment: string;
  images: string[];
  tourVisitDate: string;
  createdAt: string;
}

export interface CreateReviewRequest {
  tourId: string;
  touristId: string;
  touristName: string;
  rating: number;
  comment: string;
  images: string[];
  tourVisitDate: string;
}

@Injectable({ providedIn: 'root' })
export class ReviewService {
  private apiUrl = 'http://localhost:8080/reviews';

  constructor(private http: HttpClient) {}

  createReview(data: CreateReviewRequest): Observable<Review> {
    return this.http.post<Review>(this.apiUrl, data);
  }

  getReviewsByTour(tourId: string): Observable<Review[]> {
    return this.http.get<Review[]>(`${this.apiUrl}/tour/${tourId}`);
  }
}