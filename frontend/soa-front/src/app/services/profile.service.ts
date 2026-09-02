import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface ProfileResponse {
  username: string;
  firstName: string;
  lastName: string;
  imageURL: string;
  bio: string;
  motto: string;
}

export interface PublicProfileResponse extends ProfileResponse {
  role: string;
}

export interface UpdateProfileRequest {
  firstName?: string;
  lastName?: string;
  imageURL?: string;
  bio?: string;
  motto?: string;
}

@Injectable({ providedIn: 'root' })
export class ProfileService {
  // Auth header comes from the global authInterceptorFn - no need to attach it per call here.
  private apiUrl = `${environment.apiUrl}/stakeholders/profile`;

  constructor(private http: HttpClient) {}

  getProfile(): Observable<{ profile: ProfileResponse }> {
    return this.http.get<{ profile: ProfileResponse }>(this.apiUrl);
  }

  updateProfile(data: UpdateProfileRequest): Observable<{ profile: ProfileResponse }> {
    return this.http.put<{ profile: ProfileResponse }>(this.apiUrl, data);
  }

  getPublicProfile(username: string): Observable<{ profile: PublicProfileResponse }> {
    return this.http.get<{ profile: PublicProfileResponse }>(
      `${this.apiUrl}/${encodeURIComponent(username)}`
    );
  }

  searchProfiles(params: { username?: string; role?: string; limit?: number }):
    Observable<{ profiles: PublicProfileResponse[] }> {
    const query = new URLSearchParams();
    if (params.username) {
      query.set('username', params.username);
    }
    if (params.role) {
      query.set('role', params.role);
    }
    if (params.limit) {
      query.set('limit', String(params.limit));
    }

    const queryString = query.toString();
    const url = queryString ? `${this.apiUrl}/search?${queryString}` : `${this.apiUrl}/search`;
    return this.http.get<{ profiles: PublicProfileResponse[] }>(url);
  }
}
