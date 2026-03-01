import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { Auth } from '../../core/auth/auth';

@Component({
  selector: 'app-register',
  imports: [CommonModule, FormsModule],
  templateUrl: './register.html'
})
export class Register {
  name = '';
  email = '';
  password = '';
  loading = false;
  error = '';

  constructor(
    private readonly auth: Auth,
    private readonly router: Router
  ) {}

  onSubmit(): void {
    this.error = '';
    this.loading = true;

    this.auth.register(this.name, this.email, this.password).subscribe({
      next: () => {
        this.loading = false;
        this.router.navigate(['/projects']);
      },
      error: (err) => {
        this.loading = false;
        this.error = err?.error?.message ?? 'Registration failed. Please try again.';
      },
    });
  }
}
