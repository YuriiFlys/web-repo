import { Routes } from '@angular/router';
import { authGuard } from './core/auth/auth-guard';
import { Login } from './features/login/login';
import { ProjectDetails } from './features/project-details/project-details';
import { Projects } from './features/projects/projects';
import { Profile } from './features/profile/profile';
import { Register } from './features/register/register';

export const routes: Routes = [
  { path: 'login', component: Login },
  { path: 'register', component: Register },
  { path: 'profile', component: Profile, canActivate: [authGuard] },
  { path: 'projects', component: Projects, canActivate: [authGuard] },
  { path: 'projects/:id', component: ProjectDetails, canActivate: [authGuard] },
  { path: '', pathMatch: 'full', redirectTo: 'projects' },
  { path: '**', redirectTo: 'projects' },
];
