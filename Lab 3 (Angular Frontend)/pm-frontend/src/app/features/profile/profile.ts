import { CommonModule } from '@angular/common';
import { Component, PLATFORM_ID, inject } from '@angular/core';
import { Router } from '@angular/router';
import { Auth, AuthUser } from '../../core/auth/auth';
import { isPlatformBrowser } from '@angular/common';
import { BehaviorSubject, Observable, of } from 'rxjs';
import { catchError, map, shareReplay, startWith } from 'rxjs/operators';

@Component({
  selector: 'app-profile',
  imports: [CommonModule],
  templateUrl: './profile.html'
})
export class Profile {
  readonly user$: Observable<AuthUser | null>;
  readonly loading$: Observable<boolean>;
  readonly error$: Observable<string>;
  readonly isBrowser = isPlatformBrowser(inject(PLATFORM_ID));

  constructor(
    private readonly auth: Auth,
    private readonly router: Router
  ) {
    if (!this.isBrowser) {
      this.user$ = of(null);
      this.loading$ = of(false);
      this.error$ = of('');
      return;
    }

    const errorSubject = new BehaviorSubject<string>('');
    const request$ = this.auth.me().pipe(
      catchError(() => {
        errorSubject.next('Failed to load profile.');
        return of(null);
      }),
      shareReplay(1)
    );

    this.user$ = request$;
    this.loading$ = request$.pipe(
      startWith(true),
      catchError(() => of(null)),
      map(() => false)
    );
    this.error$ = errorSubject.asObservable();
  }

  logout(): void {
    this.auth.logout();
    this.router.navigate(['/login']);
  }

}
