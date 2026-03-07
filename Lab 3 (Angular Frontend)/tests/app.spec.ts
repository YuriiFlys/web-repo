import { expect, Page, test } from '@playwright/test';

type TaskStatus = 'todo' | 'in_progress' | 'done';

interface User {
  id: number;
  name: string;
  email: string;
  createdAt: string;
}

interface Project {
  id: number;
  title: string;
  description: string;
  status: 'active' | 'archived';
  createdAt: string;
}

interface Task {
  id: number;
  projectId: number;
  title: string;
  description: string;
  status: TaskStatus;
  assigneeId: number | null;
  createdAt: string;
}

interface Comment {
  id: number;
  taskId: number;
  author: string;
  text: string;
  createdAt: string;
}

interface MockState {
  currentUser: User;
  projects: Project[];
  tasks: Task[];
  comments: Comment[];
  users: User[];
  nextProjectId: number;
  nextTaskId: number;
  nextCommentId: number;
}

const apiBase = 'http://localhost:8080/api';

async function mockApi(page: Page): Promise<MockState> {
  const state: MockState = {
    currentUser: {
      id: 1,
      name: 'Alice Tester',
      email: 'alice@example.com',
      createdAt: '2026-03-01T09:00:00.000Z',
    },
    users: [
      {
        id: 1,
        name: 'Alice Tester',
        email: 'alice@example.com',
        createdAt: '2026-03-01T09:00:00.000Z',
      },
      {
        id: 2,
        name: 'Bob Builder',
        email: 'bob@example.com',
        createdAt: '2026-03-02T09:00:00.000Z',
      },
    ],
    projects: [
      {
        id: 101,
        title: 'Alpha Launch',
        description: 'Release planning and tracking.',
        status: 'active',
        createdAt: '2026-03-03T10:00:00.000Z',
      },
      {
        id: 102,
        title: 'Archive Cleanup',
        description: 'Historical clean-up work.',
        status: 'archived',
        createdAt: '2026-03-04T11:00:00.000Z',
      },
    ],
    tasks: [
      {
        id: 201,
        projectId: 101,
        title: 'Draft scope',
        description: 'Collect release scope and assumptions.',
        status: 'todo',
        assigneeId: 1,
        createdAt: '2026-03-03T12:00:00.000Z',
      },
      {
        id: 202,
        projectId: 101,
        title: 'QA sign-off',
        description: 'Coordinate final QA review.',
        status: 'in_progress',
        assigneeId: 2,
        createdAt: '2026-03-03T13:00:00.000Z',
      },
      {
        id: 203,
        projectId: 101,
        title: 'Publish notes',
        description: 'Prepare release notes for users.',
        status: 'done',
        assigneeId: null,
        createdAt: '2026-03-03T14:00:00.000Z',
      },
    ],
    comments: [
      {
        id: 301,
        taskId: 201,
        author: 'Alice Tester',
        text: 'Initial scope draft is ready.',
        createdAt: '2026-03-04T08:00:00.000Z',
      },
    ],
    nextProjectId: 103,
    nextTaskId: 204,
    nextCommentId: 302,
  };

  await page.route(`${apiBase}/**`, async (route) => {
    const request = route.request();
    const method = request.method();
    const url = new URL(request.url());
    const path = url.pathname.replace('/api', '');
    const bodyText = request.postData() ?? '';
    const body = bodyText ? JSON.parse(bodyText) : null;

    const json = (payload: unknown, status = 200) =>
      route.fulfill({
        status,
        contentType: 'application/json',
        body: JSON.stringify(payload),
      });

    if (method === 'POST' && path === '/auth/login') {
      return json({ token: 'mock-token', user: state.currentUser });
    }

    if (method === 'POST' && path === '/auth/register') {
      state.currentUser = {
        id: state.currentUser.id,
        name: body?.name ?? 'New User',
        email: body?.email ?? 'new@example.com',
        createdAt: state.currentUser.createdAt,
      };
      state.users[0] = state.currentUser;
      return json({ token: 'mock-token', user: state.currentUser });
    }

    if (method === 'GET' && path === '/auth/me') {
      return json(state.currentUser);
    }

    if (method === 'GET' && path === '/users') {
      return json({ items: state.users });
    }

    if (method === 'GET' && path === '/projects') {
      const q = (url.searchParams.get('q') ?? '').trim().toLowerCase();
      const status = url.searchParams.get('status') ?? '';
      const page = Number(url.searchParams.get('page') ?? '1');
      const pageSize = Number(url.searchParams.get('pageSize') ?? '20');

      let items = [...state.projects];
      if (q) {
        items = items.filter(
          (project) =>
            project.title.toLowerCase().includes(q) ||
            project.description.toLowerCase().includes(q)
        );
      }
      if (status) {
        items = items.filter((project) => project.status === status);
      }

      const start = (page - 1) * pageSize;
      const paged = items.slice(start, start + pageSize);
      return json({
        items: paged,
        page,
        pageSize,
        isLast: start + pageSize >= items.length,
      });
    }

    if (method === 'POST' && path === '/projects') {
      const project: Project = {
        id: state.nextProjectId++,
        title: body?.title ?? 'Untitled project',
        description: body?.description ?? '',
        status: body?.status ?? 'active',
        createdAt: '2026-03-05T10:00:00.000Z',
      };
      state.projects.unshift(project);
      return json(project, 201);
    }

    const projectIdMatch = path.match(/^\/projects\/(\d+)$/);
    if (projectIdMatch && method === 'GET') {
      const projectId = Number(projectIdMatch[1]);
      const project = state.projects.find((item) => item.id === projectId);
      return json(project ?? null, project ? 200 : 404);
    }

    if (projectIdMatch && method === 'DELETE') {
      const projectId = Number(projectIdMatch[1]);
      state.projects = state.projects.filter((item) => item.id !== projectId);
      state.tasks = state.tasks.filter((task) => task.projectId !== projectId);
      state.comments = state.comments.filter((comment) =>
        state.tasks.some((task) => task.id === comment.taskId)
      );
      return route.fulfill({ status: 204, body: '' });
    }

    if (method === 'GET' && path === '/tasks') {
      const projectId = Number(url.searchParams.get('projectId'));
      return json({
        items: state.tasks.filter((task) => task.projectId === projectId),
      });
    }

    if (method === 'POST' && path === '/tasks') {
      const task: Task = {
        id: state.nextTaskId++,
        projectId: body?.projectId,
        title: body?.title ?? 'Untitled task',
        description: body?.description ?? '',
        status: body?.status ?? 'todo',
        assigneeId: body?.assigneeId ?? null,
        createdAt: '2026-03-05T12:00:00.000Z',
      };
      state.tasks.unshift(task);
      return json(task, 201);
    }

    const taskIdMatch = path.match(/^\/tasks\/(\d+)$/);
    if (taskIdMatch && method === 'PUT') {
      const taskId = Number(taskIdMatch[1]);
      const index = state.tasks.findIndex((item) => item.id === taskId);
      if (index === -1) {
        return json({ message: 'Not found' }, 404);
      }
      state.tasks[index] = {
        ...state.tasks[index],
        ...body,
      };
      return json(state.tasks[index]);
    }

    if (taskIdMatch && method === 'DELETE') {
      const taskId = Number(taskIdMatch[1]);
      state.tasks = state.tasks.filter((item) => item.id !== taskId);
      state.comments = state.comments.filter((comment) => comment.taskId !== taskId);
      return route.fulfill({ status: 204, body: '' });
    }

    const taskCommentsMatch = path.match(/^\/tasks\/(\d+)\/comments$/);
    if (taskCommentsMatch && method === 'GET') {
      const taskId = Number(taskCommentsMatch[1]);
      return json({
        items: state.comments.filter((comment) => comment.taskId === taskId),
      });
    }

    if (taskCommentsMatch && method === 'POST') {
      const taskId = Number(taskCommentsMatch[1]);
      const comment: Comment = {
        id: state.nextCommentId++,
        taskId,
        author: body?.author ?? state.currentUser.name,
        text: body?.text ?? '',
        createdAt: '2026-03-05T13:00:00.000Z',
      };
      state.comments.unshift(comment);
      return json(comment, 201);
    }

    const commentIdMatch = path.match(/^\/comments\/(\d+)$/);
    if (commentIdMatch && method === 'DELETE') {
      const commentId = Number(commentIdMatch[1]);
      state.comments = state.comments.filter((item) => item.id !== commentId);
      return route.fulfill({ status: 204, body: '' });
    }

    return json({ message: `Unhandled route: ${method} ${path}` }, 500);
  });

  return state;
}

test.beforeEach(async ({ page }) => {
  await mockApi(page);
});

test('redirects unauthenticated users to login', async ({ page }) => {
  await page.goto('/projects');

  await expect(page).toHaveURL(/\/login\?returnUrl=%2Fprojects$/);
  await expect(page.getByRole('heading', { name: 'Sign in' })).toBeVisible();
});

test('registers a user and lands on the projects page', async ({ page }) => {
  await page.goto('/register');

  await page.getByLabel('Name').fill('E2E User');
  await page.getByLabel('Email').fill('e2e@example.com');
  await page.getByLabel('Password').fill('secret12');
  await page.getByRole('button', { name: 'Create account' }).click();

  await expect(page).toHaveURL(/\/projects$/);
  await expect(page.getByRole('heading', { name: 'Projects' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Profile' })).toBeVisible();
});

test('logs in, filters projects, creates a project, and deletes it', async ({ page }) => {
  page.on('dialog', (dialog) => dialog.accept());

  await page.goto('/login');
  await page.getByLabel('Email').fill('alice@example.com');
  await page.getByLabel('Password').fill('secret12');
  await page.getByRole('button', { name: 'Sign in' }).click();

  await expect(page).toHaveURL(/\/projects$/);
  await expect(page.getByText('Alpha Launch')).toBeVisible();
  await expect(page.getByText('Archive Cleanup')).toBeVisible();

  await page.getByLabel('Search').fill('archive');
  await page.getByLabel('Status').selectOption('archived');
  await page.getByRole('button', { name: 'Apply' }).click();

  await expect(page.getByText('Archive Cleanup')).toBeVisible();
  await expect(page.getByText('Alpha Launch')).not.toBeVisible();

  await page.getByRole('button', { name: 'Reset' }).click();
  await expect(page.getByText('Alpha Launch')).toBeVisible();

  await page.getByRole('button', { name: 'New project' }).click();
  await expect(page.getByRole('heading', { name: 'Create project' })).toBeVisible();
  await page.locator('input[name="newProjectTitle"]').fill('Playwright Project');
  await page.locator('textarea[name="newProjectDescription"]').fill('Created from the E2E suite.');
  await page.locator('select[name="newProjectStatus"]').selectOption('active');
  await page.getByRole('button', { name: 'Create' }).click();

  await expect(page.getByRole('status').filter({ hasText: 'Project created.' })).toBeVisible();
  await expect(page.getByText('Playwright Project')).toBeVisible();

  const projectCard = page.locator('article').filter({ hasText: 'Playwright Project' }).first();
  await projectCard.getByRole('button', { name: 'Delete' }).click();

  await expect(page.getByRole('status').filter({ hasText: 'Project deleted.' })).toBeVisible();
  await expect(page.getByText('Playwright Project')).not.toBeVisible();
});

test('manages tasks and comments from the project details board', async ({ page }) => {
  page.on('dialog', (dialog) => dialog.accept());

  await page.addInitScript(() => {
    localStorage.setItem('pm_token', 'mock-token');
    localStorage.setItem(
      'pm_user',
      JSON.stringify({
        id: 1,
        name: 'Alice Tester',
        email: 'alice@example.com',
        createdAt: '2026-03-01T09:00:00.000Z',
      })
    );
  });

  await page.goto('/projects/101');

  await expect(page.getByRole('heading', { name: 'Alpha Launch' })).toBeVisible();
  await expect(page.getByText('Draft scope')).toBeVisible();

  await page.getByRole('button', { name: 'Add task' }).click();
  await page.locator('input[name="newTitle"]').fill('Playwright Task');
  await page.locator('textarea[name="newDescription"]').fill('Task created from Playwright.');
  await page.locator('select[name="newAssigneeId"]').selectOption({
    label: 'Alice Tester (alice@example.com)',
  });
  await page.locator('select[name="newStatus"]').selectOption('done');
  await page.getByRole('button', { name: 'Create' }).click();

  await expect(page.getByRole('status').filter({ hasText: 'Task created.' })).toBeVisible();
  await expect(page.getByText('Playwright Task')).toBeVisible();

  await page.getByText('Draft scope').click();
  await expect(page.getByRole('heading', { name: 'Task details' })).toBeVisible();

  await page.locator('input[type="text"]').last().fill('Draft scope updated');
  await page.locator('input[type="text"]').last().blur();
  await expect(page.locator('input[type="text"]').last()).toHaveValue('Draft scope updated');

  await page.locator('select').first().selectOption('done');
  await expect(page.locator('select').first()).toHaveValue('done');

  await page.getByPlaceholder('Write a comment...').fill('Reviewed during E2E.');
  await page.getByRole('button', { name: 'Add comment' }).click();
  await expect(page.getByRole('status').filter({ hasText: 'Comment added.' })).toBeVisible();
  await expect(page.getByText('Reviewed during E2E.')).toBeVisible();

  await page.getByRole('button', { name: 'Delete' }).first().evaluate((element: HTMLButtonElement) => {
    element.click();
  });
  await expect(page.getByRole('status').filter({ hasText: 'Task deleted.' })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Task details' })).not.toBeVisible();
});

test('shows profile data and logs out', async ({ page }) => {
  await page.addInitScript(() => {
    localStorage.setItem('pm_token', 'mock-token');
    localStorage.setItem(
      'pm_user',
      JSON.stringify({
        id: 1,
        name: 'Alice Tester',
        email: 'alice@example.com',
        createdAt: '2026-03-01T09:00:00.000Z',
      })
    );
  });

  await page.goto('/profile');

  await expect(page.getByRole('heading', { name: 'Profile' })).toBeVisible();
  await expect(page.getByText('Alice Tester')).toBeVisible();
  await expect(page.getByText('alice@example.com')).toBeVisible();

  await page.getByRole('button', { name: 'Logout' }).click();

  await expect(page).toHaveURL(/\/login$/);
  await expect(page.getByRole('heading', { name: 'Sign in' })).toBeVisible();
});
