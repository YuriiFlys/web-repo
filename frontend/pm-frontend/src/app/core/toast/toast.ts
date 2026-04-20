import { Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';

export type ToastKind = 'success' | 'error' | 'info';

export interface Toast {
  id: number;
  kind: ToastKind;
  title: string;
  message?: string;
  timeoutMs?: number;
}

@Injectable({ providedIn: 'root' })
export class Toasts {
  private readonly toastsSubject = new BehaviorSubject<Toast[]>([]);
  readonly toasts$ = this.toastsSubject.asObservable();
  private nextId = 1;

  show(toast: Omit<Toast, 'id'>): number {
    const id = this.nextId++;
    const nextToast: Toast = { id, timeoutMs: 4000, ...toast };
    this.toastsSubject.next([nextToast, ...this.toastsSubject.value]);
    if (nextToast.timeoutMs && nextToast.timeoutMs > 0) {
      setTimeout(() => this.dismiss(id), nextToast.timeoutMs);
    }
    return id;
  }

  success(title: string, message?: string, timeoutMs = 3500): number {
    return this.show({ kind: 'success', title, message, timeoutMs });
  }

  error(title: string, message?: string, timeoutMs = 5000): number {
    return this.show({ kind: 'error', title, message, timeoutMs });
  }

  info(title: string, message?: string, timeoutMs = 4000): number {
    return this.show({ kind: 'info', title, message, timeoutMs });
  }

  dismiss(id: number): void {
    this.toastsSubject.next(this.toastsSubject.value.filter((toast) => toast.id !== id));
  }
}
