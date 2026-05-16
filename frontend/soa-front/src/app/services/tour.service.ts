import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Tour {
  id: string;
  authorId: string;
  name: string;
  description: string;
  difficulty: 'easy' | 'medium' | 'hard';
  tags: string[];
  status: string;
  price: number;
  createdAt: string;
}

export interface CreateTourRequest {
  authorId: string;
  name: string;
  description: string;
  difficulty: string;
  tags: string[];
}

export interface KeyPoint {
  id?: string;
  tourId: string;
  name: string;
  description: string;
  latitude: number;
  longitude: number;
  imageUrl: string;
  order: number;
}

@Injectable({ providedIn: 'root' })
export class TourService {
  private apiUrl = 'http://localhost:8080';

  constructor(private http: HttpClient) {}

  createTour(data: CreateTourRequest): Observable<Tour> {
    return this.http.post<Tour>(`${this.apiUrl}/tours`, data);
  }

  getToursByAuthor(authorId: string): Observable<Tour[]> {
    return this.http.get<Tour[]>(`${this.apiUrl}/tours/author/${authorId}`);
  }

  getTourById(id: string): Observable<Tour> {
    return this.http.get<Tour>(`${this.apiUrl}/tours/${id}`);
  }

  createKeyPoint(data: Omit<KeyPoint, 'id'>): Observable<KeyPoint> {
    return this.http.post<KeyPoint>(`${this.apiUrl}/keypoints`, data);
  }

  getKeyPointsByTour(tourId: string): Observable<KeyPoint[]> {
    return this.http.get<KeyPoint[]>(`${this.apiUrl}/keypoints/tour/${tourId}`);
  }

  deleteKeyPoint(id: string): Observable<any> {
    return this.http.delete(`${this.apiUrl}/keypoints/${id}`);
  }
}