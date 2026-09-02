import { Injectable, PLATFORM_ID, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { isPlatformBrowser } from '@angular/common';
import { Observable, BehaviorSubject } from 'rxjs';
import { tap } from 'rxjs/operators';
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

  logout(): void {
    if (isPlatformBrowser(this.platformId)) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
    }
    this.currentUserSubject.next(null);
  }

  getToken(): string | null {
    if (isPlatformBrowser(this.platformId)) {
      return localStorage.getItem('token');
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
