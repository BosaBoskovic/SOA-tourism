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


export const routes: Routes = [
  { path: '', redirectTo: '/login', pathMatch: 'full' },
  { path: 'login', component: LoginComponent },
  { path: 'register', component: RegisterComponent },
  { path: 'dashboard', component: DashboardComponent },
  { path: 'profile', component: ProfileComponent },
  { path: 'users/search', component: UserSearchComponent },
  { path: 'users/:username', component: UserProfileComponent },
  { path: 'tours', component: TourListComponent },
  { path: 'tours/new', component: TourCreateComponent },
  { path: 'tours/:id', component: TourDetailComponent },
  { path: 'tours/:tourId/execute', component: TourExecutionComponent },
  { path: 'tours/:tourId/execute/:executionId', component: TourExecutionComponent },
  { path: 'position-simulator', component: PositionSimulatorComponent},
  { path: 'blog', component: BlogComponent },
  { path: 'cart', component: ShoppingCartComponent },
];
