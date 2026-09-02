import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface Recommendation {
  username: string;
  score: number;
}

@Injectable({ providedIn: 'root' })
export class FollowersService {
  // Auth header comes from the global authInterceptorFn - no need to attach it per call here.
  private apiUrl = `${environment.apiUrl}/followers`;

  constructor(private http: HttpClient) {}

  follow(targetUsername: string): Observable<{ message: string; relation: any }> {
    return this.http.post<{ message: string; relation: any }>(
      `${this.apiUrl}/follow`,
      { targetUsername }
    );
  }

  unfollow(targetUsername: string): Observable<void> {
    return this.http.delete<void>(
      `${this.apiUrl}/follow/${encodeURIComponent(targetUsername)}`
    );
  }

  getFollowing(username: string): Observable<{ username: string; following: string[] }> {
    return this.http.get<{ username: string; following: string[] }>(
      `${this.apiUrl}/following/${encodeURIComponent(username)}`
    );
  }

  getFollowers(username: string): Observable<{ username: string; followers: string[] }> {
    return this.http.get<{ username: string; followers: string[] }>(
      `${this.apiUrl}/followers/${encodeURIComponent(username)}`
    );
  }

  isFollowing(followerUsername: string, targetUsername: string): Observable<{ isFollowing: boolean }> {
    const params = new HttpParams()
      .set('followerUsername', followerUsername)
      .set('targetUsername', targetUsername);

    return this.http.get<{ isFollowing: boolean }>(`${this.apiUrl}/is-following`, { params });
  }

  getRecommendations(username: string, limit = 6): Observable<{ username: string; recommendations: Recommendation[] }> {
    const params = new HttpParams().set('limit', String(limit));

    return this.http.get<{ username: string; recommendations: Recommendation[] }>(
      `${this.apiUrl}/recommendations/${encodeURIComponent(username)}`,
      { params }
    );
  }
}
