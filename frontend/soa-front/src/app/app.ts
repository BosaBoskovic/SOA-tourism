import { Component, signal } from '@angular/core';
import { NgIf } from '@angular/common';
import { NavigationEnd, Router, RouterOutlet } from '@angular/router';
import { filter } from 'rxjs/operators';
import { TopNavComponent } from './shared/top-nav/top-nav.component';

const NAV_HIDDEN_ROUTES = ['/login', '/register'];

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, TopNavComponent, NgIf],
  templateUrl: './app.html',
  styleUrl: './app.css'
})
export class App {
  protected readonly title = signal('soa-front');
  // The shell now renders the nav once, here, instead of every page
  // template opting in (or forgetting to) - hidden only on the two
  // full-screen auth pages.
  protected readonly showNav = signal(true);

  constructor(router: Router) {
    this.showNav.set(!NAV_HIDDEN_ROUTES.includes(router.url));
    router.events.pipe(filter((e): e is NavigationEnd => e instanceof NavigationEnd))
      .subscribe(event => {
        this.showNav.set(!NAV_HIDDEN_ROUTES.some(route => event.urlAfterRedirects.startsWith(route)));
      });
  }
}
