import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';

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

@Injectable({ providedIn: 'root' })
export class BlogService {
  private readonly BASE = 'http://localhost:8080/blog';

  constructor(private http: HttpClient) {}

  getAllBlogs(): Observable<BlogResponse[]> {
    return this.http.get<BlogResponse[]>(this.BASE);
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

  addComment(blogId: string, text: string): Observable<BlogResponse> {
    return this.http.post<BlogResponse>(`${this.BASE}/${blogId}/comments`, { text });
  }

  editComment(blogId: string, commentId: string, text: string): Observable<BlogResponse> {
    return this.http.put<BlogResponse>(`${this.BASE}/${blogId}/comments/${commentId}`, { text });
  }

  toggleLike(blogId: string): Observable<{ likesCount: number; likedByCurrentUser: boolean }> {
    return this.http.post<{ likesCount: number; likedByCurrentUser: boolean }>(
      `${this.BASE}/${blogId}/like`,
      {}
    );
  }
}