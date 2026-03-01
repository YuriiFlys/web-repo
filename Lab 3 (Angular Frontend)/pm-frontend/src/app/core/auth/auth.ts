import { Injectable, PLATFORM_ID, inject } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, tap } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class Auth {
  private readonly tokenKey = 'pm_token';
  private readonly userKey = 'pm_user';
  private readonly apiBase = 'http://localhost:8080/api';
  private readonly platformId = inject(PLATFORM_ID);
  private readonly isBrowser = isPlatformBrowser(this.platformId);
  private readonly authStateSubject = new BehaviorSubject<boolean>(false);
  private readonly readySubject = new BehaviorSubject<boolean>(false);

  constructor(private readonly http: HttpClient) {
    if (!this.isBrowser) {
      this.readySubject.next(true);
    }
  }

  login(email: string, password: string): Observable<AuthResponse> {
    return this.http
      .post<AuthResponse>(`${this.apiBase}/auth/login`, { email, password })
      .pipe(tap((res) => this.persist(res)));
  }

  register(name: string, email: string, password: string): Observable<AuthResponse> {
    return this.http
      .post<AuthResponse>(`${this.apiBase}/auth/register`, { name, email, password })
      .pipe(tap((res) => this.persist(res)));
  }

  me(): Observable<AuthUser> {
    return this.http.get<AuthUser>(`${this.apiBase}/auth/me`).pipe(
      tap((user) => {
        if (!this.isBrowser) {
          return;
        }
        localStorage.setItem(this.userKey, JSON.stringify(user));
      })
    );
  }

  logout(): void {
    if (!this.isBrowser) {
      return;
    }
    localStorage.removeItem(this.tokenKey);
    localStorage.removeItem(this.userKey);
    this.authStateSubject.next(false);
  }

  get token(): string | null {
    if (!this.isBrowser) {
      return null;
    }
    return localStorage.getItem(this.tokenKey);
  }

  get user(): AuthUser | null {
    if (!this.isBrowser) {
      return null;
    }
    const raw = localStorage.getItem(this.userKey);
    return raw ? (JSON.parse(raw) as AuthUser) : null;
  }

  isAuthenticated(): boolean {
    return this.authStateSubject.value;
  }

  get authState$(): Observable<boolean> {
    return this.authStateSubject.asObservable();
  }

  get ready$(): Observable<boolean> {
    return this.readySubject.asObservable();
  }

  private persist(res: AuthResponse): void {
    if (!this.isBrowser) {
      return;
    }
    localStorage.setItem(this.tokenKey, res.token);
    localStorage.setItem(this.userKey, JSON.stringify(res.user));
    this.authStateSubject.next(true);
  }

  initFromStorage(): void {
    if (!this.isBrowser) {
      return;
    }
    const hasToken = !!localStorage.getItem(this.tokenKey);
    this.authStateSubject.next(hasToken);
    this.readySubject.next(true);
  }
}

export interface AuthUser {
  id: number;
  email: string;
  name: string;
  createdAt: string;
}

export interface AuthResponse {
  token: string;
  user: AuthUser;
}
