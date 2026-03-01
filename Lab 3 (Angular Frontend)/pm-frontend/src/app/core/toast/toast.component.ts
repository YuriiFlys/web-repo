import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { Toasts, Toast } from './toast';

@Component({
  selector: 'app-toasts',
  imports: [CommonModule],
  template: `
    <section
      class="pointer-events-none fixed right-4 bottom-4 z-50 flex w-[22rem] max-w-[90vw] flex-col gap-3"
      aria-live="polite"
    >
      <article
        *ngFor="let toast of toasts.toasts$ | async; trackBy: trackById"
        class="pointer-events-auto rounded-2xl border border-slate-900/10 bg-white px-4 py-3 shadow-xl shadow-slate-900/15"
        [class.border-emerald-200]="toast.kind === 'success'"
        [class.border-rose-200]="toast.kind === 'error'"
        [class.border-slate-200]="toast.kind === 'info'"
        [class.bg-emerald-50]="toast.kind === 'success'"
        [class.bg-rose-50]="toast.kind === 'error'"
        [class.bg-slate-50]="toast.kind === 'info'"
        role="status"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <p
              class="text-sm font-semibold"
              [class.text-emerald-800]="toast.kind === 'success'"
              [class.text-rose-800]="toast.kind === 'error'"
              [class.text-slate-800]="toast.kind === 'info'"
            >
              {{ toast.title }}
            </p>
            <p class="mt-1 text-xs text-slate-600" *ngIf="toast.message">{{ toast.message }}</p>
          </div>
          <button
            type="button"
            class="rounded-full px-2 py-1 text-[10px] font-semibold uppercase tracking-widest text-slate-500 hover:bg-white/70"
            (click)="toasts.dismiss(toast.id)"
          >
            Close
          </button>
        </div>
      </article>
    </section>
  `
})
export class ToastsComponent {
  constructor(public readonly toasts: Toasts) {}

  trackById(_: number, toast: Toast): number {
    return toast.id;
  }
}
