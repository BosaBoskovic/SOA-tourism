import { Routes } from '@angular/router';
import { LoginComponent } from './auth/login/login.component';
import { RegisterComponent } from './auth/register/register.component';
import { DashboardComponent } from './dashboard/dashboard.component';
import { ProfileComponent } from './profile/profile.component';
import { TourListComponent } from './tours/tour-list/tour-list.component';
import { TourCreateComponent } from './tours/tour-create/tour-create.component';
import { TourDetailComponent } from './tours/tour-detail/tour-detail.component';
import { PositionSimulatorComponent } from './position-simulator/position-simulator.component';
import { BlogComponent } from './blog/blog.component';
import { UserSearchComponent } from './users/user-search/user-search.component';
import { UserProfileComponent } from './users/user-profile/user-profile.component';
import { ShoppingCartComponent } from './shopping-cart/shopping-cart.component';
import { TourExecutionComponent } from './tour-execution/tour-execution.component';
import { AdminComponent } from './admin/admin.component';
import { authGuard } from './auth/guards/auth.guard';


export const routes: Routes = [
  { path: '', redirectTo: '/login', pathMatch: 'full' },
  { path: 'login', component: LoginComponent },
  { path: 'register', component: RegisterComponent },
  { path: 'dashboard', component: DashboardComponent, canActivate: [authGuard] },
  { path: 'profile', component: ProfileComponent, canActivate: [authGuard] },
  { path: 'users/search', component: UserSearchComponent, canActivate: [authGuard] },
  { path: 'users/:username', component: UserProfileComponent, canActivate: [authGuard] },
  { path: 'tours', component: TourListComponent, canActivate: [authGuard] },
  { path: 'tours/new', component: TourCreateComponent, canActivate: [authGuard] },
  { path: 'tours/:id', component: TourDetailComponent, canActivate: [authGuard] },
  { path: 'tours/:tourId/execute', component: TourExecutionComponent, canActivate: [authGuard] },
  { path: 'tours/:tourId/execute/:executionId', component: TourExecutionComponent, canActivate: [authGuard] },
  { path: 'position-simulator', component: PositionSimulatorComponent, canActivate: [authGuard] },
  { path: 'blog', component: BlogComponent, canActivate: [authGuard] },
  { path: 'cart', component: ShoppingCartComponent, canActivate: [authGuard] },
  { path: 'admin', component: AdminComponent, canActivate: [authGuard] },
];
