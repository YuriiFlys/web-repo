import { CommonModule } from '@angular/common';
import { Component, PLATFORM_ID, inject } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { HttpClient } from '@angular/common/http';
import { isPlatformBrowser } from '@angular/common';
import { BehaviorSubject, of } from 'rxjs';
import { catchError, map, shareReplay } from 'rxjs/operators';
import { Comment, CommentsListResponse, Project, Task, TaskStatus, TasksListResponse, User, UsersListResponse } from '../../core/models';
import { FormsModule } from '@angular/forms';
import { Auth } from '../../core/auth/auth';
import { Toasts } from '../../core/toast/toast';

@Component({
  selector: 'app-project-details',
  imports: [CommonModule, FormsModule],
  templateUrl: './project-details.html'
})
export class ProjectDetails {
  readonly projectId: string | null;
  readonly projectIdNum: number | null;
  readonly isBrowser = isPlatformBrowser(inject(PLATFORM_ID));
  private readonly apiBase = 'http://localhost:8080/api';
  private readonly tasksSubject = new BehaviorSubject<Task[]>([]);
  private readonly commentsSubject = new BehaviorSubject<Comment[]>([]);
  private readonly usersSubject = new BehaviorSubject<User[]>([]);
  private readonly errorSubject = new BehaviorSubject<string>('');
  private readonly loadingSubject = new BehaviorSubject<boolean>(true);
  private readonly projectSubject = new BehaviorSubject<Project | null>(null);

  readonly tasks$ = this.tasksSubject.asObservable();
  readonly comments$ = this.commentsSubject.asObservable();
  readonly users$ = this.usersSubject.asObservable();
  readonly loading$ = this.loadingSubject.asObservable();
  readonly error$ = this.errorSubject.asObservable();
  readonly project$ = this.projectSubject.asObservable();

  readonly todo$ = this.tasks$.pipe(map((tasks) => tasks.filter((t) => t.status === 'todo')));
  readonly inProgress$ = this.tasks$.pipe(map((tasks) => tasks.filter((t) => t.status === 'in_progress')));
  readonly done$ = this.tasks$.pipe(map((tasks) => tasks.filter((t) => t.status === 'done')));

  showModal = false;
  createMode = false;
  selectedTask: Task | null = null;
  newTitle = '';
  newDescription = '';
  newAssigneeId: number | null = null;
  newStatus: TaskStatus = 'todo';
  newComment = '';
  private readonly pendingUpdates = new Map<string, number>();

  private readonly toasts = inject(Toasts);

  constructor(
    private readonly route: ActivatedRoute,
    private readonly http: HttpClient,
    private readonly auth: Auth
  ) {
    this.projectId = this.route.snapshot.paramMap.get('id');
    this.projectIdNum = this.projectId ? Number(this.projectId) : null;

    if (!this.isBrowser || !this.projectIdNum) {
      this.loadingSubject.next(false);
      return;
    }

    this.loadProject(String(this.projectIdNum));
    this.loadTasks(String(this.projectIdNum));
    this.loadUsers();
  }

  onDragStart(event: DragEvent, task: Task): void {
    if (!event.dataTransfer) {
      return;
    }
    event.dataTransfer.setData('text/plain', String(task.id));
    event.dataTransfer.effectAllowed = 'move';
  }

  onDrop(event: DragEvent, status: TaskStatus): void {
    event.preventDefault();
    if (!event.dataTransfer) {
      return;
    }
    const id = Number(event.dataTransfer.getData('text/plain'));
    if (!id) {
      return;
    }
    const current = this.tasksSubject.value;
    const task = current.find((t) => t.id === id);
    if (!task || task.status === status) {
      return;
    }

    const updated = current.map((t) => (t.id === id ? { ...t, status } : t));
    this.tasksSubject.next(updated);
    this.updateStatus(id, status);
  }

  allowDrop(event: DragEvent): void {
    event.preventDefault();
  }

  openDetails(task: Task): void {
    this.selectedTask = task;
    this.createMode = false;
    this.showModal = true;
    this.newComment = '';
    this.loadComments(task.id);
  }

  openCreate(): void {
    this.selectedTask = null;
    this.createMode = true;
    this.showModal = true;
    this.newTitle = '';
    this.newDescription = '';
    this.newAssigneeId = null;
    this.newStatus = 'todo';
  }

  closeModal(): void {
    this.showModal = false;
    this.createMode = false;
    this.selectedTask = null;
    this.commentsSubject.next([]);
  }

  createTask(): void {
    if (!this.projectIdNum || !this.newTitle.trim()) {
      return;
    }
    const payload = {
      projectId: this.projectIdNum,
      title: this.newTitle.trim(),
      description: this.newDescription.trim(),
      status: this.newStatus,
      assigneeId: this.newAssigneeId
    };
    this.http.post<Task>(`${this.apiBase}/tasks`, payload).pipe(
      catchError(() => {
        this.errorSubject.next('Failed to create task.');
        this.toasts?.error('Failed to create task.');
        return of(null);
      })
    ).subscribe((created) => {
      if (!created) {
        return;
      }
      this.tasksSubject.next([created, ...this.tasksSubject.value]);
      this.toasts?.success('Task created.');
      this.closeModal();
    });
  }

  addComment(): void {
    if (!this.selectedTask || !this.newComment.trim()) {
      return;
    }
    const payload = {
      author: this.getAuthorName(),
      text: this.newComment.trim()
    };
    this.http.post<Comment>(`${this.apiBase}/tasks/${this.selectedTask.id}/comments`, payload).pipe(
      catchError(() => {
        this.errorSubject.next('Failed to add comment.');
        this.toasts?.error('Failed to add comment.');
        return of(null);
      })
    ).subscribe((created) => {
      if (!created) {
        return;
      }
      this.commentsSubject.next([created, ...this.commentsSubject.value]);
      this.newComment = '';
      this.toasts?.success('Comment added.');
    });
  }

  deleteComment(comment: Comment): void {
    this.http.delete(`${this.apiBase}/comments/${comment.id}`).pipe(
      catchError(() => {
        this.errorSubject.next('Failed to delete comment.');
        this.toasts?.error('Failed to delete comment.');
        return of(null);
      })
    ).subscribe(() => {
      this.commentsSubject.next(this.commentsSubject.value.filter((c) => c.id !== comment.id));
      this.toasts?.success('Comment deleted.');
    });
  }

  deleteTask(task: Task): void {
    const confirmed = window.confirm(`Delete task "${task.title}"?`);
    if (!confirmed) {
      return;
    }
    this.http.delete(`${this.apiBase}/tasks/${task.id}`).pipe(
      catchError(() => {
        this.errorSubject.next('Failed to delete task.');
        this.toasts?.error('Failed to delete task.');
        return of(null);
      })
    ).subscribe(() => {
      this.tasksSubject.next(this.tasksSubject.value.filter((t) => t.id !== task.id));
      this.toasts?.success('Task deleted.');
      if (this.selectedTask?.id === task.id) {
        this.closeModal();
      }
    });
  }

  onFieldChange(task: Task, field: 'title' | 'description', value: string): void {
    this.applyLocalUpdate(task, field, value);
    this.scheduleUpdate(task, field, value, 1000);
  }

  onSelectChange(task: Task, field: 'status' | 'assigneeId', value: string | number | null): void {
    this.applyLocalUpdate(task, field, value);
    this.sendUpdate(task, field, value);
  }

  flushUpdate(task: Task, field: 'title' | 'description'): void {
    const key = `${task.id}:${field}`;
    const timer = this.pendingUpdates.get(key);
    if (timer) {
      clearTimeout(timer);
      this.pendingUpdates.delete(key);
      const current = this.tasksSubject.value.find((t) => t.id === task.id);
      if (current) {
        this.sendUpdate(task, field, (current as any)[field] ?? '');
      }
    }
  }

  private scheduleUpdate(
    task: Task,
    field: 'title' | 'description',
    value: string,
    delayMs: number
  ): void {
    const key = `${task.id}:${field}`;
    const existing = this.pendingUpdates.get(key);
    if (existing) {
      clearTimeout(existing);
    }
    const timer = window.setTimeout(() => {
      this.pendingUpdates.delete(key);
      this.sendUpdate(task, field, value);
    }, delayMs);
    this.pendingUpdates.set(key, timer);
  }

  private applyLocalUpdate(
    task: Task,
    field: 'title' | 'description' | 'status' | 'assigneeId',
    value: string | number | null
  ): void {
    const next = this.tasksSubject.value.map((t) =>
      t.id === task.id ? { ...t, [field]: value } : t
    );
    this.tasksSubject.next(next);
    if (this.selectedTask && this.selectedTask.id === task.id) {
      this.selectedTask = { ...this.selectedTask, [field]: value } as Task;
    }
  }

  private sendUpdate(
    task: Task,
    field: 'title' | 'description' | 'status' | 'assigneeId',
    value: string | number | null
  ): void {
    const payload: any = {};
    payload[field] = value;
    this.http.put<Task>(`${this.apiBase}/tasks/${task.id}`, payload).pipe(
      catchError(() => {
        this.errorSubject.next('Failed to update task.');
        this.toasts?.error('Failed to update task.');
        return of(null);
      })
    ).subscribe((updated) => {
      if (!updated) {
        return;
      }
      const next = this.tasksSubject.value.map((t) => (t.id === task.id ? updated : t));
      this.tasksSubject.next(next);
      if (this.selectedTask && this.selectedTask.id === task.id) {
        this.selectedTask = updated;
      }
    });
  }

  private loadTasks(projectId: string): void {
    this.loadingSubject.next(true);
    const url = `${this.apiBase}/tasks?projectId=${projectId}`;
    this.http.get<TasksListResponse>(url).pipe(
      map((res) => res.items ?? []),
      catchError(() => {
        this.errorSubject.next('Failed to load tasks.');
        this.toasts?.error('Failed to load tasks.');
        return of([]);
      }),
      shareReplay(1)
    ).subscribe((items) => {
      this.tasksSubject.next(items);
      this.loadingSubject.next(false);
    });
  }

  private loadUsers(): void {
    this.http.get<UsersListResponse>(`${this.apiBase}/users`).pipe(
      catchError(() => {
        this.errorSubject.next('Failed to load users.');
        this.toasts?.error('Failed to load users.');
        return of({ items: [] });
      })
    ).subscribe((res) => {
      this.usersSubject.next(res.items ?? []);
    });
  }

  private loadComments(taskId: number): void {
    this.http.get<CommentsListResponse>(`${this.apiBase}/tasks/${taskId}/comments`).pipe(
      catchError(() => {
        this.errorSubject.next('Failed to load comments.');
        this.toasts?.error('Failed to load comments.');
        return of({ items: [] });
      })
    ).subscribe((res) => {
      this.commentsSubject.next(res.items ?? []);
    });
  }

  private getAuthorName(): string {
    const current = this.auth.user;
    if (current?.name) {
      return current.name;
    }
    const fallback = this.usersSubject.value[0];
    return fallback ? fallback.name : 'Unknown';
  }

  private loadProject(projectId: string): void {
    this.http.get<Project>(`${this.apiBase}/projects/${projectId}`).pipe(
      catchError(() => {
        this.errorSubject.next('Failed to load project.');
        this.toasts?.error('Failed to load project.');
        return of(null);
      })
    ).subscribe((project) => {
      this.projectSubject.next(project);
    });
  }

  private updateStatus(id: number, status: TaskStatus): void {
    this.http.put<Task>(`${this.apiBase}/tasks/${id}`, { status }).pipe(
      catchError(() => {
        this.errorSubject.next('Failed to update task.');
        this.toasts?.error('Failed to update task.');
        return of(null);
      })
    ).subscribe();
  }

  getAssigneeName(assigneeId?: number | null): string {
    if (!assigneeId) {
      return 'Unassigned';
    }
    const user = this.usersSubject.value.find((u) => u.id === assigneeId);
    return user ? user.name : `User #${assigneeId}`;
  }
}
