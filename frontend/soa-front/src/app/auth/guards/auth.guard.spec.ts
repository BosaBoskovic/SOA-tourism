import { TestBed } from '@angular/core/testing';
import { Router, UrlTree, provideRouter } from '@angular/router';
import { authGuard } from './auth.guard';
import { AuthService } from '../services/auth.service';

describe('authGuard', () => {
  let isLoggedIn: boolean;

  beforeEach(() => {
    isLoggedIn = false;
    TestBed.configureTestingModule({
      providers: [
        provideRouter([]),
        { provide: AuthService, useValue: { isLoggedIn: () => isLoggedIn } },
      ],
    });
  });

  function runGuard(url: string) {
    return TestBed.runInInjectionContext(() =>
      authGuard({} as any, { url } as any)
    );
  }

  it('allows navigation when the user is logged in', () => {
    isLoggedIn = true;
    expect(runGuard('/tours')).toBe(true);
  });

  it('redirects to /login with a returnUrl when the user is logged out', () => {
    isLoggedIn = false;
    const result = runGuard('/tours');

    expect(result).toBeInstanceOf(UrlTree);
    const router = TestBed.inject(Router);
    expect(router.serializeUrl(result as UrlTree)).toBe('/login?returnUrl=%2Ftours');
  });
});
