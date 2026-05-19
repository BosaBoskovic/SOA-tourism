import { Component, OnInit, ChangeDetectorRef, NgZone } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { AuthService } from '../../auth/services/auth.service';
import { FollowersService, Recommendation } from '../../services/followers.service';
import { ProfileService, PublicProfileResponse } from '../../services/profile.service';
import { TopNavComponent } from '../../shared/top-nav/top-nav.component';

@Component({
  selector: 'app-user-search',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink, TopNavComponent],
  templateUrl: './user-search.component.html',
  styleUrl: './user-search.component.css'
})
export class UserSearchComponent implements OnInit {
  searchForm!: FormGroup;

  profiles: PublicProfileResponse[] = [];
  recommendations: Recommendation[] = [];
  following = new Set<string>();

  currentUser: any = null;
  loading = false;
  recommendationsLoading = false;
  error = '';

  constructor(
    private profileService: ProfileService,
    private followersService: FollowersService,
    private authService: AuthService,
    private fb: FormBuilder,
    private cdr: ChangeDetectorRef,
    private zone: NgZone
  ) {
    this.searchForm = this.fb.group({
      username: [''],
      role: ['']
    });
  }

  ngOnInit(): void {
    this.authService.currentUser$.subscribe(user => {
      this.currentUser = user;
      if (user?.username) {
        this.loadFollowing(user.username);
        this.loadRecommendations(user.username);
      }
    });

    this.loadInitialProfiles();
  }

  search(): void {
    const username = (this.searchForm.value.username || '').trim();
    const role = (this.searchForm.value.role || '').trim();

    if (!username && !role) {
      this.error = 'Unesi korisnicko ime ili izaberi ulogu.';
      this.profiles = [];
      return;
    }

    this.loading = true;
    this.error = '';

    this.profileService.searchProfiles({ username, role, limit: 25 }).subscribe({
      next: (res) => {
        this.zone.run(() => {
          this.profiles = res.profiles || [];
          this.loading = false;
          this.cdr.detectChanges();
        });
      },
      error: (err: any) => {
        this.zone.run(() => {
          this.error = err.error?.error || 'Greska pri pretrazi korisnika.';
          this.loading = false;
          this.cdr.detectChanges();
        });
      }
    });
  }

  clearSearch(): void {
    this.searchForm.reset({ username: '', role: '' });
    this.profiles = [];
    this.error = '';
    this.loadInitialProfiles();
  }

  isFollowing(username: string): boolean {
    return this.following.has(username);
  }

  canFollow(username: string): boolean {
    return !!this.currentUser && username !== this.currentUser.username;
  }

  followUser(username: string): void {
    if (!this.canFollow(username)) {
      return;
    }

    this.followersService.follow(username).subscribe({
      next: () => {
        this.zone.run(() => {
          this.following.add(username);
          if (this.currentUser?.username) {
            this.loadRecommendations(this.currentUser.username);
          }
          this.cdr.detectChanges();
        });
      },
      error: (err: any) => {
        this.zone.run(() => {
          this.error = err.error?.error || 'Greska pri pracenju korisnika.';
          this.cdr.detectChanges();
        });
      }
    });
  }

  unfollowUser(username: string): void {
    if (!this.canFollow(username)) {
      return;
    }

    this.followersService.unfollow(username).subscribe({
      next: () => {
        this.zone.run(() => {
          this.following.delete(username);
          if (this.currentUser?.username) {
            this.loadRecommendations(this.currentUser.username);
          }
          this.cdr.detectChanges();
        });
      },
      error: (err: any) => {
        this.zone.run(() => {
          this.error = err.error?.error || 'Greska pri otpracenju korisnika.';
          this.cdr.detectChanges();
        });
      }
    });
  }

  private loadFollowing(username: string): void {
    this.followersService.getFollowing(username).subscribe({
      next: (res) => {
        this.zone.run(() => {
          this.following = new Set(res.following || []);
          this.cdr.detectChanges();
        });
      },
      error: () => {
        this.zone.run(() => {
          this.following = new Set();
          this.cdr.detectChanges();
        });
      }
    });
  }

  private loadRecommendations(username: string): void {
    this.recommendationsLoading = true;

    this.followersService.getRecommendations(username, 6).subscribe({
      next: (res) => {
        this.zone.run(() => {
          this.recommendations = res.recommendations || [];
          this.recommendationsLoading = false;
          this.cdr.detectChanges();
        });
      },
      error: () => {
        this.zone.run(() => {
          this.recommendations = [];
          this.recommendationsLoading = false;
          this.cdr.detectChanges();
        });
      }
    });
  }

  private loadInitialProfiles(): void {
    this.loading = true;
    this.error = '';

    this.profileService.searchProfiles({ limit: 12 }).subscribe({
      next: (res) => {
        this.zone.run(() => {
          this.profiles = res.profiles || [];
          this.loading = false;
          this.cdr.detectChanges();
        });
      },
      error: () => {
        this.zone.run(() => {
          this.profiles = [];
          this.loading = false;
          this.cdr.detectChanges();
        });
      }
    });
  }
}
