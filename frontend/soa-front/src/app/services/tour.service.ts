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
  lengthKm: number;
  durations: TourDuration[];
  price: number;
  createdAt: string;
  updatedAt?: string;
  publishedAt?: string;
  archivedAt?: string;
}

export interface TourPreview {
  id: string;
  authorId: string;
  name: string;
  description: string;
  difficulty: 'easy' | 'medium' | 'hard';
  tags: string[];
  lengthKm: number;
  price: number;
  publishedAt?: string;
  firstKeyPoint?: KeyPoint;
}

export type TransportType = 'walk' | 'bike' | 'car';

export interface TourDuration {
  transport: TransportType;
  minutes: number;
}

export interface CreateTourRequest {
  authorId: string;
  name: string;
  description: string;
  difficulty: string;
  tags: string[];
  durations?: TourDuration[];
}

export interface UpdateTourRequest {
  name: string;
  description: string;
  difficulty: string;
  tags: string[];
  durations: TourDuration[];
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

  updateTour(id: string, data: UpdateTourRequest): Observable<Tour> {
    return this.http.put<Tour>(`${this.apiUrl}/tours/${id}`, data);
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

  getAllTours(): Observable<TourPreview[]> {
    return this.http.get<TourPreview[]>(`${this.apiUrl}/tours`);
  }

  publishTour(tourId: string): Observable<Tour> {
    return this.http.put<Tour>(`${this.apiUrl}/tours/${tourId}/publish`, {});
  }

  archiveTour(tourId: string): Observable<Tour> {
    return this.http.put<Tour>(`${this.apiUrl}/tours/${tourId}/archive`, {});
  }

  activateTour(tourId: string): Observable<Tour> {
    return this.http.put<Tour>(`${this.apiUrl}/tours/${tourId}/activate`, {});
  }

  updateKeyPoint(id: string, data: Omit<KeyPoint, 'id'>): Observable<KeyPoint> {
    return this.http.put<KeyPoint>(`${this.apiUrl}/keypoints/${id}`, data);
  }
}