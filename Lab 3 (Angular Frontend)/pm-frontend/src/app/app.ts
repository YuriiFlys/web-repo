import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { Auth } from './core/auth/auth';
import { ToastsComponent } from './core/toast/toast.component';

@Component({
  selector: 'app-root',
  imports: [CommonModule, RouterOutlet, RouterLink, RouterLinkActive, ToastsComponent],
  templateUrl: './app.html'
})
export class App {
  constructor(public readonly auth: Auth) {
    this.auth.initFromStorage();
  }
}
