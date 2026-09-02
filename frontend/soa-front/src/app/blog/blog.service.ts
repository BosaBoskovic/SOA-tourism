import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface Comment {
  id: string;
  authorUsername: string;
  text: string;
  createdAt: string;
  lastModifiedAt?: string;
}

export interface BlogData {
  id: string;
  title: string;
  descriptionMarkdown: string;
  authorUsername: string;
  createdAt: string;
  imageUrls: string[];
  comments: Comment[];
}

export interface BlogResponse {
  blog: BlogData;
  likesCount: number;
  likedByCurrentUser: boolean;
  descriptionHtml?: string;
}

export interface BlogListResponse {
  blogs: BlogResponse[];
  page: number;
  size: number;
  totalElements: number;
  totalPages: number;
}

@Injectable({ providedIn: 'root' })
export class BlogService {
  private readonly BASE = `${environment.apiUrl}/blog`;

  constructor(private http: HttpClient) {}

  getAllBlogs(page = 0, size = 10): Observable<BlogListResponse> {
    return this.http.get<BlogListResponse>(`${this.BASE}?page=${page}&size=${size}`);
  }

  getBlogById(id: string): Observable<BlogResponse> {
    return this.http.get<BlogResponse>(`${this.BASE}/${id}`);
  }

  createBlog(payload: {
    title: string;
    descriptionMarkdown: string;
    imageUrls: string[];
  }): Observable<BlogResponse> {
    return this.http.post<BlogResponse>(this.BASE, payload);
  }

  updateBlog(id: string, payload: {
    title: string;
    descriptionMarkdown: string;
    imageUrls: string[];
  }): Observable<BlogResponse> {
    return this.http.put<BlogResponse>(`${this.BASE}/${id}`, payload);
  }

  deleteBlog(id: string): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(`${this.BASE}/${id}`);
  }

  addComment(blogId: string, text: string): Observable<BlogResponse> {
    return this.http.post<BlogResponse>(`${this.BASE}/${blogId}/comments`, { text });
  }

  editComment(blogId: string, commentId: string, text: string): Observable<BlogResponse> {
    return this.http.put<BlogResponse>(`${this.BASE}/${blogId}/comments/${commentId}`, { text });
  }

  deleteComment(blogId: string, commentId: string): Observable<BlogResponse> {
    return this.http.delete<BlogResponse>(`${this.BASE}/${blogId}/comments/${commentId}`);
  }

  toggleLike(blogId: string): Observable<{ likesCount: number; likedByCurrentUser: boolean }> {
    return this.http.post<{ likesCount: number; likedByCurrentUser: boolean }>(
      `${this.BASE}/${blogId}/like`,
      {}
    );
  }
}
