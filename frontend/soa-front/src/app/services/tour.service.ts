import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

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
  status?: string;
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
  price: number;
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

export interface KeyPointRequest extends Omit<KeyPoint, 'id'> {
  lengthKm?: number;
}

export interface TourDetailResponse {
  tour: Tour;
  keyPoints: KeyPoint[];
  purchased: boolean;
}

export interface TourSearchParams {
  difficulty?: string;
  tags?: string[];
  minPrice?: number;
  maxPrice?: number;
  minLengthKm?: number;
  maxLengthKm?: number;
  sortBy?: 'price' | 'length' | 'name';
  sortDir?: 'asc' | 'desc';
}

export interface TourAnalytics {
  tourId: string;
  purchaseCount: number;
  revenue: number;
}

@Injectable({ providedIn: 'root' })
export class TourService {
  private apiUrl = environment.apiUrl;

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

  createKeyPoint(data: KeyPointRequest): Observable<KeyPoint> {
    return this.http.post<KeyPoint>(`${this.apiUrl}/keypoints`, data);
  }

  getKeyPointsByTour(tourId: string): Observable<KeyPoint[]> {
    return this.http.get<KeyPoint[]>(`${this.apiUrl}/keypoints/tour/${tourId}`);
  }

  deleteKeyPoint(id: string, lengthKm?: number): Observable<any> {
    const url = lengthKm === undefined
      ? `${this.apiUrl}/keypoints/${id}`
      : `${this.apiUrl}/keypoints/${id}?lengthKm=${lengthKm}`;
    return this.http.delete(url);
  }

  getAllTours(params?: TourSearchParams): Observable<TourPreview[]> {
    const query = new URLSearchParams();
    if (params?.difficulty) query.set('difficulty', params.difficulty);
    if (params?.tags?.length) query.set('tags', params.tags.join(','));
    if (params?.minPrice != null) query.set('minPrice', String(params.minPrice));
    if (params?.maxPrice != null) query.set('maxPrice', String(params.maxPrice));
    if (params?.minLengthKm != null) query.set('minLengthKm', String(params.minLengthKm));
    if (params?.maxLengthKm != null) query.set('maxLengthKm', String(params.maxLengthKm));
    if (params?.sortBy) query.set('sortBy', params.sortBy);
    if (params?.sortDir) query.set('sortDir', params.sortDir);

    const queryString = query.toString();
    const url = queryString ? `${this.apiUrl}/tours?${queryString}` : `${this.apiUrl}/tours`;
    return this.http.get<TourPreview[]>(url);
  }

  // Guide-facing purchase/revenue analytics, aggregated by payments.
  getAnalytics(tourIds: string[]): Observable<TourAnalytics[]> {
    if (tourIds.length === 0) {
      return new Observable(subscriber => { subscriber.next([]); subscriber.complete(); });
    }
    return this.http.post<TourAnalytics[]>(`${environment.apiUrl}/checkout/analytics`, tourIds);
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

  updateKeyPoint(id: string, data: KeyPointRequest): Observable<KeyPoint> {
    return this.http.put<KeyPoint>(`${this.apiUrl}/keypoints/${id}`, data);
  }

  getTourByIdForTourist(id: string, touristId: string): Observable<TourDetailResponse> {
    return this.http.get<TourDetailResponse>(`${this.apiUrl}/tours/${id}?touristId=${touristId}`);
  }
}