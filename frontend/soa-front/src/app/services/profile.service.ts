import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface ProfileResponse {
  username: string;
  firstName: string;
  lastName: string;
  imageURL: string;
  bio: string;
  motto: string;
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
  private apiUrl = 'http://localhost:8080/stakeholders/profile';

  constructor(private http: HttpClient) {}

  getProfile(): Observable<{ profile: ProfileResponse }> {
    return this.http.get<{ profile: ProfileResponse }>(this.apiUrl);
  }

  updateProfile(data: UpdateProfileRequest): Observable<{ profile: ProfileResponse }> {
    return this.http.put<{ profile: ProfileResponse }>(this.apiUrl, data);
  }
}