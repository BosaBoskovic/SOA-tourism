import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { environment } from '../../environments/environment';

// Uploads a file to the gateway's /uploads endpoint and returns a full URL
// (environment.apiUrl + the relative path the backend hands back), replacing
// the old base64-inline-JSON pattern used across profile/keypoint/review/blog images.
@Injectable({ providedIn: 'root' })
export class UploadService {
  private apiUrl = `${environment.apiUrl}/uploads`;

  constructor(private http: HttpClient) {}

  upload(file: File): Observable<string> {
    const formData = new FormData();
    formData.append('file', file);
    return this.http.post<{ url: string }>(this.apiUrl, formData).pipe(
      map(res => `${environment.apiUrl}${res.url}`)
    );
  }
}
