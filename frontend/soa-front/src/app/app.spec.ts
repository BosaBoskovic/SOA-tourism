import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { of } from 'rxjs';
import { App } from './app';

// The old default spec asserted an <h1> with "Hello, soa-front" - text
// that hasn't existed since the top-nav shell replaced the CLI scaffold.
// These tests exercise what App actually does now: deciding whether to
// show the top nav based on the current URL. A fake Router (rather than
// provideRouter + real navigation) keeps this focused on App's own logic
// without needing TopNavComponent and its own service tree to render.
describe('App', () => {
  async function setup(initialUrl: string) {
    const fakeRouter = { url: initialUrl, events: of() } as unknown as Router;
    await TestBed.configureTestingModule({
      imports: [App],
      providers: [{ provide: Router, useValue: fakeRouter }],
    }).compileComponents();
    return TestBed.createComponent(App);
  }

  it('should create the app', async () => {
    const fixture = await setup('/dashboard');
    expect(fixture.componentInstance).toBeTruthy();
  });

  it('shows the nav outside the full-screen auth pages', async () => {
    const fixture = await setup('/dashboard');
    // showNav is protected on App - reading it via the instance (rather
    // than rendering <app-top-nav> and its own dependencies) keeps this a
    // focused test of the shell's show/hide decision.
    expect((fixture.componentInstance as any).showNav()).toBe(true);
  });

  it('hides the nav on /login', async () => {
    const fixture = await setup('/login');
    expect((fixture.componentInstance as any).showNav()).toBe(false);
  });

  it('hides the nav on /register', async () => {
    const fixture = await setup('/register');
    expect((fixture.componentInstance as any).showNav()).toBe(false);
  });
});
