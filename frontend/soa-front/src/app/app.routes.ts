import { Routes } from '@angular/router';
import { LoginComponent } from './auth/login/login.component';
import { RegisterComponent } from './auth/register/register.component';
import { DashboardComponent } from './dashboard/dashboard.component';
import { ProfileComponent } from './profile/profile.component';
import { TourListComponent } from './tours/tour-list/tour-list.component';
import { TourCreateComponent } from './tours/tour-create/tour-create.component';
import { TourDetailComponent } from './tours/tour-detail/tour-detail.component';

export const routes: Routes = [
  { path: '', redirectTo: '/login', pathMatch: 'full' },
  { path: 'login', component: LoginComponent },
  { path: 'register', component: RegisterComponent },
  { path: 'dashboard', component: DashboardComponent },
  { path: 'profile', component: ProfileComponent },
  { path: 'tours', component: TourListComponent },
  { path: 'tours/new', component: TourCreateComponent },
  { path: 'tours/:id', component: TourDetailComponent },
];
