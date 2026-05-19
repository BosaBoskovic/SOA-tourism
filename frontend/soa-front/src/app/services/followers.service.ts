import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Recommendation {
  username: string;
  score: number;
}

@Injectable({ providedIn: 'root' })
export class FollowersService {
  private apiUrl = 'http://localhost:8080/followers';

  constructor(private http: HttpClient) {}

  private getAuthHeaders(): HttpHeaders {
    const token = localStorage.getItem('token');

    return new HttpHeaders({
      Authorization: `Bearer ${token}`
    });
  }

  follow(targetUsername: string): Observable<{ message: string; relation: any }> {
    return this.http.post<{ message: string; relation: any }>(
      `${this.apiUrl}/follow`,
      { targetUsername },
      { headers: this.getAuthHeaders() }
    );
  }

  unfollow(targetUsername: string): Observable<void> {
    return this.http.delete<void>(
      `${this.apiUrl}/follow/${encodeURIComponent(targetUsername)}`,
      { headers: this.getAuthHeaders() }
    );
  }

  getFollowing(username: string): Observable<{ username: string; following: string[] }> {
    return this.http.get<{ username: string; following: string[] }>(
      `${this.apiUrl}/following/${encodeURIComponent(username)}`,
      { headers: this.getAuthHeaders() }
    );
  }

  isFollowing(followerUsername: string, targetUsername: string): Observable<{ isFollowing: boolean }> {
    const params = new HttpParams()
      .set('followerUsername', followerUsername)
      .set('targetUsername', targetUsername);

    return this.http.get<{ isFollowing: boolean }>(
      `${this.apiUrl}/is-following`,
      { params, headers: this.getAuthHeaders() }
    );
  }

  getRecommendations(username: string, limit = 6): Observable<{ username: string; recommendations: Recommendation[] }> {
    const params = new HttpParams().set('limit', String(limit));

    return this.http.get<{ username: string; recommendations: Recommendation[] }>(
      `${this.apiUrl}/recommendations/${encodeURIComponent(username)}`,
      { params, headers: this.getAuthHeaders() }
    );
  }
}
