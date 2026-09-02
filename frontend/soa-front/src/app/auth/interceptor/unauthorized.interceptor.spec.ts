import { TestBed } from '@angular/core/testing';
import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';
import { unauthorizedInterceptorFn } from './unauthorized.interceptor';
import { AuthService, AuthResponse } from '../services/auth.service';

const API = 'http://localhost:8080';

describe('unauthorizedInterceptorFn', () => {
  let http: HttpClient;
  let httpMock: HttpTestingController;
  let authServiceStub: {
    getRefreshToken: () => string | null;
    isLoggedIn: () => boolean;
    refreshAccessToken: () => any;
    logout: () => void;
  };

  beforeEach(() => {
    authServiceStub = {
      getRefreshToken: () => 'refresh-abc',
      isLoggedIn: () => true,
      refreshAccessToken: () =>
        of({
          accessToken: 'new-token',
          message: 'ok',
          account: { username: 'ana', email: 'ana@example.com', role: 'tourist' },
        } as AuthResponse),
      logout: () => {},
    };

    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([unauthorizedInterceptorFn])),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: AuthService, useValue: authServiceStub },
      ],
    });

    http = TestBed.inject(HttpClient);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('silently refreshes on a 401 and retries the original request with the new token', () => {
    let result: any;
    http.get(`${API}/tours`).subscribe(res => (result = res));

    const first = httpMock.expectOne(`${API}/tours`);
    first.flush('unauthorized', { status: 401, statusText: 'Unauthorized' });

    const retried = httpMock.expectOne(`${API}/tours`);
    expect(retried.request.headers.get('Authorization')).toBe('Bearer new-token');
    retried.flush({ ok: true });

    expect(result).toEqual({ ok: true });
  });

  it('does not attempt a refresh for the login endpoint itself, and just propagates the 401', () => {
    // If the interceptor ever mistakenly tried to refresh here, this throw
    // would surface as the observed error instead of a plain 401 - proof
    // the auth-endpoint exclusion actually short-circuits before refreshing.
    authServiceStub.refreshAccessToken = () => {
      throw new Error('refreshAccessToken should not be called for an auth endpoint');
    };

    let sawError: any;
    http.post(`${API}/stakeholders/login`, {}).subscribe({ error: err => (sawError = err) });

    httpMock.expectOne(`${API}/stakeholders/login`)
      .flush('bad credentials', { status: 401, statusText: 'Unauthorized' });

    expect(sawError.status).toBe(401);
  });
});
