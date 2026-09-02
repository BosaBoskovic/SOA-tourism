import { Injectable, PLATFORM_ID, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { isPlatformBrowser } from '@angular/common';
import { Observable, BehaviorSubject, throwError } from 'rxjs';
import { catchError, tap } from 'rxjs/operators';
import { environment } from '../../../environments/environment';

export interface LoginRequest {
  usernameOrEmail: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
  email: string;
  role: string;
}

export interface AuthAccount {
  username: string;
  email: string;
  role: string;
}

export interface AuthResponse {
  accessToken?: string;
  refreshToken?: string;
  message: string;
  account: AuthAccount;
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private apiGatewayUrl = environment.apiUrl;
  private currentUserSubject = new BehaviorSubject<AuthAccount | null>(null);
  public currentUser$ = this.currentUserSubject.asObservable();

  constructor(
    private http: HttpClient,
    @Inject(PLATFORM_ID) private platformId: Object
  ) {
    this.loadUser();
  }

  login(credentials: LoginRequest): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiGatewayUrl}/stakeholders/login`, credentials)
      .pipe(tap(response => this.applySession(response)));
  }

  register(data: RegisterRequest): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiGatewayUrl}/stakeholders/register`, data)
      .pipe(tap(response => this.applySession(response)));
  }

  // Exchanges the stored refresh token for a new access token. Used by
  // unauthorizedInterceptor to keep a session alive silently instead of
  // forcing a full re-login every 15 minutes.
  refreshAccessToken(): Observable<AuthResponse> {
    const refreshToken = this.getRefreshToken();
    if (!refreshToken) {
      return throwError(() => new Error('No refresh token available'));
    }
    return this.http.post<AuthResponse>(`${this.apiGatewayUrl}/stakeholders/refresh`, { refreshToken })
      .pipe(
        tap(response => this.applySession(response)),
        catchError(err => {
          this.logout();
          return throwError(() => err);
        })
      );
  }

  changePassword(currentPassword: string, newPassword: string): Observable<{ message: string }> {
    return this.http.put<{ message: string }>(`${this.apiGatewayUrl}/stakeholders/password`, {
      currentPassword,
      newPassword,
    });
  }

  // No email service exists anywhere in this stack - the response carries
  // the reset token/link directly instead of it being emailed. That's a
  // demo-only shortcut; a real deployment must remove resetToken from the
  // response and actually send an email.
  requestPasswordReset(usernameOrEmail: string): Observable<{ message: string; resetToken?: string }> {
    return this.http.post<{ message: string; resetToken?: string }>(
      `${this.apiGatewayUrl}/stakeholders/password-reset/request`,
      { usernameOrEmail }
    );
  }

  confirmPasswordReset(token: string, newPassword: string): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(`${this.apiGatewayUrl}/stakeholders/password-reset/confirm`, {
      token,
      newPassword,
    });
  }

  logout(): void {
    const refreshToken = this.getRefreshToken();
    if (isPlatformBrowser(this.platformId)) {
      localStorage.removeItem('token');
      localStorage.removeItem('refreshToken');
      localStorage.removeItem('user');
    }
    this.currentUserSubject.next(null);
    if (refreshToken) {
      // Best-effort - the client-side session is already cleared either way.
      this.http.post(`${this.apiGatewayUrl}/stakeholders/logout`, { refreshToken }).subscribe({ error: () => {} });
    }
  }

  getToken(): string | null {
    if (isPlatformBrowser(this.platformId)) {
      return localStorage.getItem('token');
    }
    return null;
  }

  getRefreshToken(): string | null {
    if (isPlatformBrowser(this.platformId)) {
      return localStorage.getItem('refreshToken');
    }
    return null;
  }

  isLoggedIn(): boolean {
    return !!this.getToken();
  }

  getCurrentUser(): AuthAccount | null {
    return this.currentUserSubject.getValue();
  }

  // A response with no accessToken (shouldn't normally happen, but the
  // backend contract does mark it optional) must NOT leave the app
  // believing the user is logged in - every subsequent API call would go
  // out with no Authorization header and just fail.
  private applySession(response: AuthResponse): void {
    if (!response.accessToken) {
      return;
    }
    if (isPlatformBrowser(this.platformId)) {
      localStorage.setItem('token', response.accessToken);
      if (response.refreshToken) {
        localStorage.setItem('refreshToken', response.refreshToken);
      }
      localStorage.setItem('user', JSON.stringify(response.account));
    }
    this.currentUserSubject.next(response.account);
  }

  private loadUser(): void {
    if (isPlatformBrowser(this.platformId)) {
      // A user record with no token is not a logged-in session.
      if (!localStorage.getItem('token')) {
        return;
      }
      const userJson = localStorage.getItem('user');
      if (userJson) {
        try {
          this.currentUserSubject.next(JSON.parse(userJson));
        } catch (e) {
          localStorage.removeItem('user');
        }
      }
    }
  }
}
