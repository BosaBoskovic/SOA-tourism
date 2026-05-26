import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../auth/services/auth.service';
import { CartService } from '../../services/cart.service';

@Component({
  selector: 'app-top-nav',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './top-nav.component.html',
  styleUrl: './top-nav.component.css'
})
export class TopNavComponent implements OnInit {
  currentUser: any = null;
  menuOpen = false;
  cartItemCount = 0;

  constructor(
    private authService: AuthService,
    private cartService: CartService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.authService.currentUser$.subscribe(user => {
      this.currentUser = user;
      if (user?.role === 'tourist') {
        this.cartService.getCart(user.username).subscribe({
          next: (cart) => { this.cartItemCount = cart.items?.length ?? 0; },
          error: () => { this.cartItemCount = 0; }
        });
      }
    });

    this.cartService.cart$.subscribe(cart => {
      this.cartItemCount = cart?.items?.length ?? 0;
    });
  }

  toggleMenu(): void {
    this.menuOpen = !this.menuOpen;
  }

  logout(): void {
    this.authService.logout();
    this.router.navigate(['/login']);
  }
}