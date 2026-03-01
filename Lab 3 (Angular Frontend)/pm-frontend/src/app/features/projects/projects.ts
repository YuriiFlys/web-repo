import { CommonModule } from '@angular/common';
import { Component, PLATFORM_ID, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { isPlatformBrowser } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { ProjectsListResponse, Project } from '../../core/models';
import { BehaviorSubject, Observable, of } from 'rxjs';
import { catchError, map, shareReplay, startWith, switchMap } from 'rxjs/operators';
import { Toasts } from '../../core/toast/toast';

@Component({
  selector: 'app-projects',
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './projects.html'
})
export class Projects {
  readonly projects$: Observable<Project[]>;
  readonly loading$: Observable<boolean>;
  readonly error$: Observable<string>;
  readonly isLast$: Observable<boolean>;
  private readonly apiBase = 'http://localhost:8080/api';
  readonly isBrowser = isPlatformBrowser(inject(PLATFORM_ID));
  private readonly errorSubject = new BehaviorSubject<string>('');
  private readonly filtersSubject = new BehaviorSubject<ProjectFilters>({
    q: '',
    status: '',
    page: 1,
    pageSize: 20
  });

  readonly filters$ = this.filtersSubject.asObservable();
  q = '';
  status = '';
  page = 1;
  pageSize = 20;
  showCreate = false;
  newTitle = '';
  newDescription = '';
  newStatus = 'active';

  private readonly toasts = inject(Toasts);

  constructor(
    private readonly http: HttpClient
  ) {
    if (!this.isBrowser) {
      this.projects$ = of([]);
      this.isLast$ = of(true);
      this.loading$ = of(false);
      this.error$ = of('');
      return;
    }

    const response$ = this.filtersSubject.pipe(
      switchMap((filters) => {
        const params = new URLSearchParams();
        if (filters.q) {
          params.set('q', filters.q);
        }
        if (filters.status) {
          params.set('status', filters.status);
        }
        params.set('page', String(filters.page));
        params.set('pageSize', String(filters.pageSize));
        const url = params.toString()
          ? `${this.apiBase}/projects?${params.toString()}`
          : `${this.apiBase}/projects`;

        return this.http.get<ProjectsListResponse>(url).pipe(
          catchError(() => {
            this.errorSubject.next('Failed to load projects.');
            this.toasts?.error('Failed to load projects.');
            return of({ items: [], isLast: true } as ProjectsListResponse);
          })
        );
      }),
      shareReplay(1)
    );

    this.projects$ = response$.pipe(map((res) => res.items ?? []));
    this.isLast$ = response$.pipe(map((res) => !!res.isLast));
    this.loading$ = response$.pipe(
      map(() => false),
      startWith(true)
    );
    this.error$ = this.errorSubject.asObservable();
  }

  applyFilters(): void {
    this.page = 1;
    this.filtersSubject.next({
      q: this.q.trim(),
      status: this.status,
      page: this.page,
      pageSize: this.pageSize
    });
  }

  clearFilters(): void {
    this.q = '';
    this.status = '';
    this.page = 1;
    this.filtersSubject.next({ q: '', status: '', page: this.page, pageSize: this.pageSize });
  }

  nextPage(): void {
    this.page += 1;
    this.filtersSubject.next({
      q: this.q.trim(),
      status: this.status,
      page: this.page,
      pageSize: this.pageSize
    });
  }

  prevPage(): void {
    if (this.page <= 1) {
      return;
    }
    this.page -= 1;
    this.filtersSubject.next({
      q: this.q.trim(),
      status: this.status,
      page: this.page,
      pageSize: this.pageSize
    });
  }

  openCreate(): void {
    this.showCreate = true;
    this.newTitle = '';
    this.newDescription = '';
    this.newStatus = 'active';
    this.errorSubject.next('');
  }

  closeCreate(): void {
    this.showCreate = false;
    this.errorSubject.next('');
  }

  createProject(): void {
    if (!this.newTitle.trim()) {
      this.errorSubject.next('Project title is required.');
      this.toasts?.error('Project title is required.');
      return;
    }
    const payload = {
      title: this.newTitle.trim(),
      description: this.newDescription.trim(),
      status: this.newStatus
    };
    this.http.post<Project>(`${this.apiBase}/projects`, payload).pipe(
      catchError(() => {
        this.errorSubject.next('Failed to create project.');
        this.toasts?.error('Failed to create project.');
        return of(null);
      })
    ).subscribe((created) => {
      if (!created) {
        return;
      }
      this.errorSubject.next('');
      this.showCreate = false;
      this.toasts?.success('Project created.');
      this.page = 1;
      this.filtersSubject.next({
        q: this.q.trim(),
        status: this.status,
        page: this.page,
        pageSize: this.pageSize
      });
    });
  }

  deleteProject(project: Project, event?: Event): void {
    if (event) {
      event.preventDefault();
      event.stopPropagation();
    }
    const confirmed = window.confirm(`Delete project "${project.title}"?`);
    if (!confirmed) {
      return;
    }
    this.http.delete(`${this.apiBase}/projects/${project.id}`).pipe(
      catchError(() => {
        this.errorSubject.next('Failed to delete project.');
        this.toasts?.error('Failed to delete project.');
        return of(null);
      })
    ).subscribe(() => {
      this.toasts?.success('Project deleted.');
      this.page = 1;
      this.filtersSubject.next({
        q: this.q.trim(),
        status: this.status,
        page: this.page,
        pageSize: this.pageSize
      });
    });
  }

}

interface ProjectFilters {
  q: string;
  status: string;
  page: number;
  pageSize: number;
}
