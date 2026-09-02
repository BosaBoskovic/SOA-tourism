import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { AuthService, AuthResponse } from './auth.service';
import { environment } from '../../../environments/environment';

const fakeLoginResponse: AuthResponse = {
  accessToken: 'access-123',
  refreshToken: 'refresh-456',
  message: 'ok',
  account: { username: 'ana', email: 'ana@example.com', role: 'tourist' },
};

describe('AuthService', () => {
  let service: AuthService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(AuthService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
    localStorage.clear();
  });

  it('starts logged out when localStorage has no token', () => {
    expect(service.isLoggedIn()).toBe(false);
    expect(service.getCurrentUser()).toBeNull();
  });

  it('login() stores the token pair and publishes the account on success', () => {
    let received: AuthResponse | undefined;
    service.login({ usernameOrEmail: 'ana', password: 'password123' }).subscribe(res => (received = res));

    const req = httpMock.expectOne(`${environment.apiUrl}/stakeholders/login`);
    expect(req.request.method).toBe('POST');
    req.flush(fakeLoginResponse);

    expect(received).toEqual(fakeLoginResponse);
    expect(service.isLoggedIn()).toBe(true);
    expect(localStorage.getItem('token')).toBe('access-123');
    expect(localStorage.getItem('refreshToken')).toBe('refresh-456');
    expect(service.getCurrentUser()?.username).toBe('ana');
  });

  it('does not treat a response with no accessToken as a login', () => {
    service.login({ usernameOrEmail: 'ana', password: 'password123' }).subscribe();

    const req = httpMock.expectOne(`${environment.apiUrl}/stakeholders/login`);
    req.flush({ message: 'pending', account: fakeLoginResponse.account } as AuthResponse);

    expect(service.isLoggedIn()).toBe(false);
    expect(localStorage.getItem('token')).toBeNull();
  });

  it('logout() clears the session and best-effort notifies the backend when a refresh token existed', () => {
    service.login({ usernameOrEmail: 'ana', password: 'password123' }).subscribe();
    httpMock.expectOne(`${environment.apiUrl}/stakeholders/login`).flush(fakeLoginResponse);

    service.logout();

    expect(service.isLoggedIn()).toBe(false);
    expect(service.getCurrentUser()).toBeNull();
    expect(localStorage.getItem('token')).toBeNull();
    expect(localStorage.getItem('refreshToken')).toBeNull();

    const logoutReq = httpMock.expectOne(`${environment.apiUrl}/stakeholders/logout`);
    expect(logoutReq.request.body).toEqual({ refreshToken: 'refresh-456' });
    logoutReq.flush({});
  });

  it('logout() with no prior session clears local state without calling the backend', () => {
    service.logout();
    expect(service.isLoggedIn()).toBe(false);
    httpMock.expectNone(`${environment.apiUrl}/stakeholders/logout`);
  });
});
