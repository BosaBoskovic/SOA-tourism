import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface Encounter {
  id: string;
  name: string;
  description: string;
  tourId?: string;
  keyPointId?: string;
  latitude: number;
  longitude: number;
  radiusMeters: number;
  reward: string;
  createdBy: string;
  createdAt: string;
  claimed: boolean;
  completedAt?: string;
}

export interface CreateEncounterRequest {
  name: string;
  description: string;
  tourId?: string;
  keyPointId?: string;
  latitude: number;
  longitude: number;
  radiusMeters: number;
  reward: string;
}

@Injectable({ providedIn: 'root' })
export class EncountersService {
  private apiUrl = `${environment.apiUrl}/encounters`;

  constructor(private http: HttpClient) {}

  list(): Observable<{ encounters: Encounter[] }> {
    return this.http.get<{ encounters: Encounter[] }>(this.apiUrl);
  }

  mine(): Observable<{ encounters: Encounter[] }> {
    return this.http.get<{ encounters: Encounter[] }>(`${this.apiUrl}/mine`);
  }

  create(req: CreateEncounterRequest): Observable<Encounter> {
    return this.http.post<Encounter>(this.apiUrl, req);
  }

  claim(id: string, latitude: number, longitude: number): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/${id}/claim`, { latitude, longitude });
  }

  delete(id: string): Observable<any> {
    return this.http.delete<any>(`${this.apiUrl}/${id}`);
  }
}
