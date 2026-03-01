import { inject, PLATFORM_ID } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { isPlatformBrowser } from '@angular/common';
import { Auth } from './auth';

export const authGuard: CanActivateFn = (route, state) => {
  const platformId = inject(PLATFORM_ID);
  if (!isPlatformBrowser(platformId)) {
    return true;
  }
  const auth = inject(Auth);
  if (auth.isAuthenticated() || !!auth.token) {
    return true;
  }
  const router = inject(Router);
  return router.createUrlTree(['/login'], { queryParams: { returnUrl: state.url } });
};
