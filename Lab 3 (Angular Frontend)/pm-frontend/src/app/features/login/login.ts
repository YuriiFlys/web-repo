import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { FormsModule, NgForm } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { Auth } from '../../core/auth/auth';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-login',
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './login.html'
})
export class Login {
  email = '';
  password = '';
  loading = false;
  error = '';
  submitted = false;

  constructor(
    private readonly auth: Auth,
    private readonly router: Router,
    private readonly route: ActivatedRoute
  ) {}

  onSubmit(form: NgForm): void {
    this.submitted = true;
    this.error = '';
    if (form.invalid) {
      this.error = 'Please fix the highlighted fields.';
      return;
    }
    this.loading = true;

    this.auth.login(this.email, this.password).subscribe({
      next: () => {
        this.loading = false;
        this.error = '';
        const returnUrl = this.route.snapshot.queryParamMap.get('returnUrl') ?? '/projects';
        this.router.navigateByUrl(returnUrl);
      },
      error: (err) => {
        this.loading = false;
        this.error = err?.error?.message ?? 'Login failed. Please try again.';
      },
    });
  }
}
