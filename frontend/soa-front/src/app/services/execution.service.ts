import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export type ExecutionStatus = 'active' | 'completed' | 'abandoned';

export interface GeoPoint {
  latitude: number;
  longitude: number;
}

export interface CompletedKeyPoint {
  keyPointId: string;
  reachedAt: string;
}

export interface TourExecution {
  id: string;
  touristId: string;
  tourId: string;
  status: ExecutionStatus;
  startLocation: GeoPoint;
  completedKeyPoints: CompletedKeyPoint[];
  startedAt: string;
  lastActivityAt: string;
  completedAt?: string;
  abandonedAt?: string;
}

export interface StartExecutionRequest {
  touristId: string;
  tourId: string;
  latitude: number;
  longitude: number;
}

export interface CheckKeyPointResponse {
  keyPointReached: boolean;
  keyPoint?: {
    id: string;
    name: string;
    description: string;
    latitude: number;
    longitude: number;
    order: number;
  };
  execution: TourExecution;
}

@Injectable({ providedIn: 'root' })
export class ExecutionService {
  private apiUrl = 'http://localhost:8080';

  constructor(private http: HttpClient) {}

  // POST /executions
  start(req: StartExecutionRequest): Observable<TourExecution> {
    return this.http.post<TourExecution>(`${this.apiUrl}/executions`, req);
  }

  // PUT /tourist-position
updateTouristPosition(req: {
  touristId: string;
  latitude: number;
  longitude: number;
}): Observable<any> {
  return this.http.put<any>(`${this.apiUrl}/tourist-position`, req);
}

  // GET /executions/{id}
  getById(id: string): Observable<TourExecution> {
    return this.http.get<TourExecution>(`${this.apiUrl}/executions/${id}`);
  }

  // GET /executions/tourist/{touristId}
  getByTourist(touristId: string): Observable<TourExecution[]> {
    return this.http.get<TourExecution[]>(`${this.apiUrl}/executions/tourist/${touristId}`);
  }

  // POST /executions/{id}/check-keypoint
  checkKeyPoint(executionId: string, latitude: number, longitude: number): Observable<CheckKeyPointResponse> {
    return this.http.post<CheckKeyPointResponse>(
      `${this.apiUrl}/executions/${executionId}/check-keypoint`,
      { latitude, longitude }
    );
  }

  // PUT /executions/{id}/complete
  complete(executionId: string): Observable<TourExecution> {
    return this.http.put<TourExecution>(`${this.apiUrl}/executions/${executionId}/complete`, {});
  }

  // PUT /executions/{id}/abandon
  abandon(executionId: string): Observable<TourExecution> {
    return this.http.put<TourExecution>(`${this.apiUrl}/executions/${executionId}/abandon`, {});
  }
}