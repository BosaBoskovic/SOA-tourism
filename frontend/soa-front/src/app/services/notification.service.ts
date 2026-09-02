import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface AppNotification {
  id: string;
  type: 'follow' | 'comment' | 'review';
  message: string;
  relatedUsername?: string;
  createdAt: string;
  read: boolean;
}

@Injectable({ providedIn: 'root' })
export class NotificationService {
  private apiUrl = `${environment.apiUrl}/stakeholders/notifications`;

  constructor(private http: HttpClient) {}

  list(): Observable<{ notifications: AppNotification[] }> {
    return this.http.get<{ notifications: AppNotification[] }>(this.apiUrl);
  }

  markRead(id: string): Observable<any> {
    return this.http.patch<any>(`${this.apiUrl}/${id}/read`, {});
  }

  markAllRead(): Observable<any> {
    return this.http.patch<any>(`${this.apiUrl}/read-all`, {});
  }
}
