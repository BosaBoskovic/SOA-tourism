import { HttpErrorResponse, HttpInterceptorFn, HttpRequest } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { BehaviorSubject, catchError, filter, switchMap, take, throwError } from 'rxjs';
import { AuthService } from '../services/auth.service';

// Endpoints a 401 from means "wrong credentials"/"expired reset token" etc,
// not "session expired" - retrying those with a refreshed access token
// makes no sense.
const AUTH_ENDPOINTS = ['/stakeholders/login', '/stakeholders/register', '/stakeholders/refresh'];

let isRefreshing = false;
const refreshedToken$ = new BehaviorSubject<string | null>(null);

function withAuthHeader(req: HttpRequest<unknown>, token: string): HttpRequest<unknown> {
  return req.clone({ setHeaders: { Authorization: `Bearer ${token}` } });
}

// A 401 first tries a silent refresh (see AuthService.refreshAccessToken) so
// an expired 15-minute access token doesn't interrupt the user - only a
// failed refresh (or no refresh token at all) clears the session and sends
// them back to /login.
export const unauthorizedInterceptorFn: HttpInterceptorFn = (req, next) => {
  const authService = inject(AuthService);
  const router = inject(Router);

  return next(req).pipe(
    catchError((error: unknown) => {
      const isAuthEndpoint = AUTH_ENDPOINTS.some(path => req.url.includes(path));
      if (!(error instanceof HttpErrorResponse) || error.status !== 401 || isAuthEndpoint || !authService.getRefreshToken()) {
        if (error instanceof HttpErrorResponse && error.status === 401 && authService.isLoggedIn() && !isAuthEndpoint) {
          authService.logout();
          router.navigate(['/login'], { queryParams: { returnUrl: router.url } });
        }
        return throwError(() => error);
      }

      if (!isRefreshing) {
        isRefreshing = true;
        refreshedToken$.next(null);

        return authService.refreshAccessToken().pipe(
          switchMap(response => {
            isRefreshing = false;
            refreshedToken$.next(response.accessToken ?? null);
            return next(withAuthHeader(req, response.accessToken!));
          }),
          catchError(refreshErr => {
            isRefreshing = false;
            router.navigate(['/login'], { queryParams: { returnUrl: router.url } });
            return throwError(() => refreshErr);
          })
        );
      }

      // A refresh is already in flight (e.g. two requests 401'd at once) -
      // wait for it instead of firing a second refresh call.
      return refreshedToken$.pipe(
        filter(token => token !== null),
        take(1),
        switchMap(token => next(withAuthHeader(req, token!)))
      );
    })
  );
};
