export type TaskStatus = 'todo' | 'in_progress' | 'done';

export interface Task {
  id: number;
  projectId: number;
  title: string;
  description?: string;
  status: TaskStatus;
  assigneeId?: number | null;
  dueDate?: string | null;
  createdAt: string;
}

export interface TasksListResponse {
  items: Task[];
}

export interface Comment {
  id: number;
  taskId: number;
  author: string;
  text: string;
  createdAt: string;
}

export interface CommentsListResponse {
  items: Comment[];
}

export interface Project {
  id: number;
  title: string;
  description: string;
  status: string;
  createdAt: string;
}

export interface ProjectsListResponse {
  items: Project[];
  page?: number;
  pageSize?: number;
  isLast?: boolean;
}

export interface User {
  id: number;
  name: string;
  email: string;
}

export interface UsersListResponse {
  items: User[];
}
