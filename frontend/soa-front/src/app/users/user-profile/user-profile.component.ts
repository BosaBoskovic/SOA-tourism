import { Component, OnInit, ChangeDetectorRef, NgZone } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { AuthService } from '../../auth/services/auth.service';
import { FollowersService } from '../../services/followers.service';
import { ProfileService, PublicProfileResponse } from '../../services/profile.service';

@Component({
  selector: 'app-user-profile',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './user-profile.component.html',
  styleUrl: './user-profile.component.css'
})
export class UserProfileComponent implements OnInit {
  profile: PublicProfileResponse | null = null;
  currentUser: any = null;

  loading = false;
  error = '';
  isFollowing = false;
  isSelf = false;
  updating = false;

  activeTab: 'following' | 'followers' = 'following';
  followingList: string[] = [];
  followersList: string[] = [];
  listsLoaded = false;

  constructor(
    private route: ActivatedRoute,
    private profileService: ProfileService,
    private followersService: FollowersService,
    private authService: AuthService,
    private cdr: ChangeDetectorRef,
    private zone: NgZone
  ) {}

  ngOnInit(): void {
    this.authService.currentUser$.subscribe(user => {
      this.currentUser = user;
    });

    this.route.paramMap.subscribe(params => {
      const username = params.get('username') || '';
      if (username) {
        this.loadProfile(username);
      }
    });
  }

  loadProfile(username: string): void {
    this.loading = true;
    this.error = '';

    this.profileService.getPublicProfile(username).subscribe({
      next: (res) => {
        this.zone.run(() => {
          this.profile = res.profile;
          this.isSelf = this.currentUser?.username === res.profile.username;
          this.loading = false;
          this.cdr.detectChanges();
          if (!this.isSelf && this.currentUser?.username) {
            this.checkFollowStatus(this.currentUser.username, res.profile.username);
          }
          this.loadLists(res.profile.username);
        });
      },
      error: (err) => {
        this.zone.run(() => {
          this.error = err.error?.error || 'Greska pri ucitavanju profila.';
          this.loading = false;
          this.cdr.detectChanges();
        });
      }
    });
  }

  toggleFollow(): void {
    if (!this.profile || this.isSelf || this.updating) {
      return;
    }

    this.updating = true;

    if (this.isFollowing) {
      this.followersService.unfollow(this.profile.username).subscribe({
        next: () => {
          this.zone.run(() => {
            this.isFollowing = false;
            this.updating = false;
            this.cdr.detectChanges();
          });
        },
        error: (err: any) => {
          this.zone.run(() => {
            this.error = err.error?.error || 'Greska pri azuriranju pracenja.';
            this.updating = false;
            this.cdr.detectChanges();
          });
        }
      });
      return;
    }

    this.followersService.follow(this.profile.username).subscribe({
      next: () => {
        this.zone.run(() => {
          this.isFollowing = true;
          this.updating = false;
          this.cdr.detectChanges();
        });
      },
      error: (err: any) => {
        this.zone.run(() => {
          this.error = err.error?.error || 'Greska pri azuriranju pracenja.';
          this.updating = false;
          this.cdr.detectChanges();
        });
      }
    });
  }

  setTab(tab: 'following' | 'followers'): void {
    this.activeTab = tab;
  }

  private loadLists(username: string): void {
    this.listsLoaded = false;
    this.followersService.getFollowing(username).subscribe({
      next: (res) => {
        this.zone.run(() => {
          this.followingList = res.following;
          this.listsLoaded = true;
          this.cdr.detectChanges();
        });
      },
      error: () => { this.zone.run(() => { this.listsLoaded = true; this.cdr.detectChanges(); }); }
    });
    this.followersService.getFollowers(username).subscribe({
      next: (res) => {
        this.zone.run(() => {
          this.followersList = res.followers;
          this.cdr.detectChanges();
        });
      },
      error: () => {}
    });
  }

  private checkFollowStatus(followerUsername: string, targetUsername: string): void {
    this.followersService.isFollowing(followerUsername, targetUsername).subscribe({
      next: (res) => {
        this.zone.run(() => {
          this.isFollowing = !!res.isFollowing;
          this.cdr.detectChanges();
        });
      },
      error: () => {
        this.zone.run(() => {
          this.isFollowing = false;
          this.cdr.detectChanges();
        });
      }
    });
  }
}
