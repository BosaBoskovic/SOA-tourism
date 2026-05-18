import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';
import { BlogService, BlogResponse, Comment } from './blog.service';
import { AuthService } from '../auth/services/auth.service';

@Component({
  selector: 'app-blog',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './blog.component.html',
  styleUrl: './blog.component.css'
})
export class BlogComponent implements OnInit {
    currentUser: { username: string; role: string } | null = null;

  // ── State
  blogs: BlogResponse[] = [];
  selectedBlog: BlogResponse | null = null;
  isLoading = false;
  error: string | null = null;

  // ── Detail state
  activeImageIndex = 0;
  isLiking = false;

  // ── Comment state
  newCommentText = '';
  isSubmittingComment = false;
  commentError: string | null = null;
  editingCommentId: string | null = null;
  editCommentText = '';

  // ── Create modal state
  showCreateModal = false;
  mdTab: 'write' | 'preview' = 'write';
  newBlog = { title: '', descriptionMarkdown: '' };
  imageUrlsRaw = '';
  isCreating = false;
  createError: string | null = null;

  constructor(
    private blogService: BlogService,
    private router: Router,
    private sanitizer: DomSanitizer,
    private cdr: ChangeDetectorRef,
    private authService: AuthService
  ) {}

  ngOnInit(): void {
    this.authService.currentUser$.subscribe(user => {
      this.currentUser = user;
      this.cdr.detectChanges();
    });
    this.loadBlogs();
  }

  loadBlogs(): void {
    this.isLoading = true;
    this.error = null;
    this.cdr.detectChanges();
    this.blogService.getAllBlogs().subscribe({
      next: (data) => {
        this.blogs = data;
        this.isLoading = false;
        this.cdr.detectChanges();
      },
      error: (err) => {
        this.error = err.status === 403
          ? 'Nemaš pristup blogovima. Provjeri da li pratiš korisnike.'
          : 'Greška pri učitavanju blogova. Provjeri da li je server aktivan.';
        this.isLoading = false;
        this.cdr.detectChanges();
      }
    });
  }

  openBlog(item: BlogResponse): void {
    this.isLoading = true;
    this.blogService.getBlogById(item.blog.id).subscribe({
      next: (data) => {
        this.selectedBlog = data;
        this.activeImageIndex = 0;
        this.isLoading = false;
        window.scrollTo(0, 0);
      },
      error: () => {
        // fallback: prikaži bez HTML renderinga
        this.selectedBlog = item;
        this.isLoading = false;
      }
    });
  }

  closeBlog(): void {
    this.selectedBlog = null;
    this.commentError = null;
    this.editingCommentId = null;
  }

  goBack(): void {
    this.router.navigate(['/dashboard']);
  }

  // ── LIKES ─────────────────────────────────────────────────────────
  toggleLike(): void {
    if (!this.selectedBlog || this.isLiking) return;
    this.isLiking = true;
    this.blogService.toggleLike(this.selectedBlog.blog.id).subscribe({
      next: (res) => {
        if (this.selectedBlog) {
          this.selectedBlog.likesCount = res.likesCount;
          this.selectedBlog.likedByCurrentUser = res.likedByCurrentUser;
          // sinhronizuj i listu
          const idx = this.blogs.findIndex(b => b.blog.id === this.selectedBlog!.blog.id);
          if (idx !== -1) {
            this.blogs[idx].likesCount = res.likesCount;
            this.blogs[idx].likedByCurrentUser = res.likedByCurrentUser;
          }
        }
        this.isLiking = false;
      },
      error: () => { this.isLiking = false; }
    });
  }

  // ── KOMENTARI ─────────────────────────────────────────────────────
  submitComment(): void {
    if (!this.selectedBlog || !this.newCommentText.trim()) return;
    this.isSubmittingComment = true;
    this.commentError = null;
    this.blogService.addComment(this.selectedBlog.blog.id, this.newCommentText.trim()).subscribe({
      next: (res) => {
        this.selectedBlog!.blog.comments = res.blog.comments;
        this.newCommentText = '';
        this.isSubmittingComment = false;
      },
      error: (err) => {
        this.commentError = err.status === 403
          ? 'Moraš pratiti autora da bi komentarisao/la.'
          : 'Greška pri slanju komentara.';
        this.isSubmittingComment = false;
      }
    });
  }

  startEdit(comment: Comment): void {
    this.editingCommentId = comment.id;
    this.editCommentText = comment.text;
  }

  cancelEdit(): void {
    this.editingCommentId = null;
    this.editCommentText = '';
  }

  saveEdit(commentId: string): void {
    if (!this.selectedBlog || !this.editCommentText.trim()) return;
    this.blogService.editComment(
      this.selectedBlog.blog.id,
      commentId,
      this.editCommentText.trim()
    ).subscribe({
      next: (res) => {
        this.selectedBlog!.blog.comments = res.blog.comments;
        this.editingCommentId = null;
      },
      error: () => {
        this.commentError = 'Greška pri izmjeni komentara.';
      }
    });
  }

  // ── KREIRANJE ─────────────────────────────────────────────────────
  openCreateModal(): void {
    this.showCreateModal = true;
    this.newBlog = { title: '', descriptionMarkdown: '' };
    this.imageUrlsRaw = '';
    this.createError = null;
    this.mdTab = 'write';
  }

  closeCreateModal(): void {
    this.showCreateModal = false;
  }

  createBlog(): void {
    if (!this.newBlog.title.trim()) return;
    this.isCreating = true;
    this.createError = null;

    const imageUrls = this.imageUrlsRaw
      .split('\n')
      .map(u => u.trim())
      .filter(u => u.length > 0);

    this.blogService.createBlog({
      title: this.newBlog.title.trim(),
      descriptionMarkdown: this.newBlog.descriptionMarkdown,
      imageUrls
    }).subscribe({
      next: (res) => {
        this.blogs.unshift(res);
        this.isCreating = false;
        this.closeCreateModal();
      },
      error: () => {
        this.createError = 'Greška pri kreiranju bloga.';
        this.isCreating = false;
      }
    });
  }

  // ── HELPERS ───────────────────────────────────────────────────────
  getExcerpt(md: string): string {
    if (!md) return '';
    // Strip markdown syntax for preview
    const plain = md
      .replace(/#{1,6}\s/g, '')
      .replace(/\*\*|__|\*|_|~~|`/g, '')
      .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
      .replace(/!\[[^\]]*\]\([^)]+\)/g, '')
      .replace(/^\s*[-*+>\d.]\s/gm, '')
      .trim();
    return plain.length > 160 ? plain.slice(0, 160) + '...' : plain;
  }

  /**
   * Lightweight client-side Markdown → HTML renderer.
   * Backend renderuje pravi CommonMark, ali ovo služi za preview u modalu
   * i fallback kada nema descriptionHtml iz API-ja.
   */
  markdownToHtml(md: string): SafeHtml {
    if (!md) return this.sanitizer.bypassSecurityTrustHtml('');

    let html = md
      // Headings
      .replace(/^#{6}\s(.+)$/gm, '<h6>$1</h6>')
      .replace(/^#{5}\s(.+)$/gm, '<h5>$1</h5>')
      .replace(/^#{4}\s(.+)$/gm, '<h4>$1</h4>')
      .replace(/^#{3}\s(.+)$/gm, '<h3>$1</h3>')
      .replace(/^#{2}\s(.+)$/gm, '<h2>$1</h2>')
      .replace(/^#{1}\s(.+)$/gm, '<h1>$1</h1>')
      // Bold & italic
      .replace(/\*\*\*(.+?)\*\*\*/g, '<strong><em>$1</em></strong>')
      .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.+?)\*/g, '<em>$1</em>')
      .replace(/__(.+?)__/g, '<strong>$1</strong>')
      .replace(/_(.+?)_/g, '<em>$1</em>')
      // Strikethrough
      .replace(/~~(.+?)~~/g, '<del>$1</del>')
      // Code inline
      .replace(/`([^`]+)`/g, '<code>$1</code>')
      // Blockquote
      .replace(/^>\s(.+)$/gm, '<blockquote>$1</blockquote>')
      // Horizontal rule
      .replace(/^---$/gm, '<hr>')
      // Images
      .replace(/!\[([^\]]*)\]\(([^)]+)\)/g, '<img src="$2" alt="$1" style="max-width:100%;border-radius:8px;margin:8px 0;">')
      // Links
      .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>')
      // Unordered list items
      .replace(/^[-*+]\s(.+)$/gm, '<li>$1</li>')
      // Ordered list items
      .replace(/^\d+\.\s(.+)$/gm, '<li>$1</li>')
      // Wrap consecutive <li> in <ul>
      .replace(/(<li>.+<\/li>\n?)+/g, (m) => `<ul>${m}</ul>`)
      // Paragraphs: double newlines
      .replace(/\n\n/g, '</p><p>')
      // Single newlines → <br>
      .replace(/([^>])\n([^<])/g, '$1<br>$2');

    html = `<p>${html}</p>`;
    // Clean up empty paragraphs
    html = html.replace(/<p><\/p>/g, '').replace(/<p>(<h[1-6]>)/g, '$1').replace(/(<\/h[1-6]>)<\/p>/g, '$1');

    return this.sanitizer.bypassSecurityTrustHtml(html);
  }
}