import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';
import { BlogService, BlogResponse, Comment } from './blog.service';
import { AuthService } from '../auth/services/auth.service';
import { ConfirmDialogService } from '../shared/confirm-dialog/confirm-dialog.service';
import { ToastService } from '../shared/toast/toast.service';

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

  // ── Create/edit modal state
  showCreateModal = false;
  editingBlogId: string | null = null;
  mdTab: 'write' | 'preview' = 'write';
  newBlog = { title: '', descriptionMarkdown: '' };
  imageUrlsRaw = '';
  isCreating = false;
  createError: string | null = null;

  // ── Pagination
  page = 0;
  pageSize = 10;
  totalPages = 0;
  totalElements = 0;

  constructor(
    private blogService: BlogService,
    private router: Router,
    private sanitizer: DomSanitizer,
    private cdr: ChangeDetectorRef,
    private authService: AuthService,
    private confirmDialogService: ConfirmDialogService,
    private toastService: ToastService
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
    this.blogService.getAllBlogs(this.page, this.pageSize).subscribe({
      next: (res) => {
        this.blogs = res.blogs;
        this.totalPages = res.totalPages;
        this.totalElements = res.totalElements;
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

  goToPage(page: number): void {
    if (page < 0 || page >= this.totalPages || page === this.page) return;
    this.page = page;
    this.loadBlogs();
    window.scrollTo(0, 0);
  }

  openBlog(item: BlogResponse): void {
    this.isLoading = true;
    this.cdr.detectChanges();
    this.blogService.getBlogById(item.blog.id).subscribe({
      next: (data) => {
        this.selectedBlog = data;
        this.activeImageIndex = 0;
        this.isLoading = false;
        window.scrollTo(0, 0);
        this.cdr.detectChanges();
      },
      error: () => {
        this.selectedBlog = item;
        this.isLoading = false;
        this.cdr.detectChanges();
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
        this.cdr.detectChanges();
      },
      error: () => { this.isLiking = false; this.cdr.detectChanges(); }
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
        this.cdr.detectChanges();
      },
      error: (err) => {
        this.commentError = err.status === 403
          ? 'Moraš pratiti autora da bi komentarisao/la.'
          : 'Greška pri slanju komentara.';
        this.isSubmittingComment = false;
        this.cdr.detectChanges();
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
        this.cdr.detectChanges();
      },
      error: () => {
        this.commentError = 'Greška pri izmjeni komentara.';
        this.cdr.detectChanges();
      }
    });
  }

  // Autor komentara ILI autor bloga smiju obrisati komentar (isto pravilo kao na backendu).
  canDeleteComment(comment: Comment): boolean {
    if (!this.currentUser || !this.selectedBlog) return false;
    return comment.authorUsername === this.currentUser.username
      || this.selectedBlog.blog.authorUsername === this.currentUser.username;
  }

  async deleteComment(comment: Comment): Promise<void> {
    if (!this.selectedBlog || !this.canDeleteComment(comment)) return;
    const confirmed = await this.confirmDialogService.confirm(
      'Da li sigurno želiš da obrišeš ovaj komentar?',
      { confirmLabel: 'Obriši' }
    );
    if (!confirmed) return;

    this.blogService.deleteComment(this.selectedBlog.blog.id, comment.id).subscribe({
      next: (res) => {
        this.selectedBlog!.blog.comments = res.blog.comments;
        this.cdr.detectChanges();
      },
      error: () => this.toastService.error('Greška pri brisanju komentara.')
    });
  }

  // ── KREIRANJE / IZMJENA (dijele isti modal) ──────────────────────
  openCreateModal(): void {
    this.editingBlogId = null;
    this.showCreateModal = true;
    this.newBlog = { title: '', descriptionMarkdown: '' };
    this.imageUrlsRaw = '';
    this.createError = null;
    this.mdTab = 'write';
  }

  closeCreateModal(): void {
    this.showCreateModal = false;
    this.editingBlogId = null;
  }

  // Poziva se iz forme u modalu - prosljeđuje na create ili edit u zavisnosti
  // od toga da li je modal otvoren za novi post ili izmjenu postojećeg.
  submitBlogForm(): void {
    if (this.editingBlogId) {
      this.saveBlogEdit();
    } else {
      this.createBlog();
    }
  }

  createBlog(): void {
    if (!this.newBlog.title.trim()) return;
    this.isCreating = true;
    this.createError = null;

    const imageUrls = this.imageUrlsRaw
      .split('\n')
      .map(u => u.trim())
      .filter(u => u.length > 0 && this.isValidUrl(u)); // filtriramo nevažeće URL-ove

    this.blogService.createBlog({
      title: this.newBlog.title.trim(),
      descriptionMarkdown: this.newBlog.descriptionMarkdown,
      imageUrls
    }).subscribe({
      next: () => {
        this.isCreating = false;
        this.closeCreateModal();
        this.page = 0;
        this.loadBlogs(); // osvježi listu (i totalElements/totalPages) umjesto ručnog prepend-a
      },
      error: () => {
        this.createError = 'Greška pri kreiranju bloga.';
        this.isCreating = false;
        this.cdr.detectChanges();
      }
    });
  }

  // ── IZMJENA / BRISANJE BLOGA ─────────────────────────────────────
  isAuthor(item: BlogResponse | null): boolean {
    return !!item && !!this.currentUser && item.blog.authorUsername === this.currentUser.username;
  }

  openEditModal(item: BlogResponse): void {
    this.editingBlogId = item.blog.id;
    this.newBlog = { title: item.blog.title, descriptionMarkdown: item.blog.descriptionMarkdown };
    this.imageUrlsRaw = (item.blog.imageUrls || []).join('\n');
    this.createError = null;
    this.mdTab = 'write';
    this.showCreateModal = true;
  }

  saveBlogEdit(): void {
    if (!this.editingBlogId || !this.newBlog.title.trim()) return;
    this.isCreating = true;
    this.createError = null;

    const imageUrls = this.imageUrlsRaw
      .split('\n')
      .map(u => u.trim())
      .filter(u => u.length > 0 && this.isValidUrl(u));

    this.blogService.updateBlog(this.editingBlogId, {
      title: this.newBlog.title.trim(),
      descriptionMarkdown: this.newBlog.descriptionMarkdown,
      imageUrls
    }).subscribe({
      next: (res) => {
        this.isCreating = false;
        this.closeCreateModal();
        if (this.selectedBlog && this.selectedBlog.blog.id === res.blog.id) {
          this.selectedBlog = res;
        }
        this.loadBlogs();
      },
      error: () => {
        this.createError = 'Greška pri izmjeni bloga.';
        this.isCreating = false;
        this.cdr.detectChanges();
      }
    });
  }

  async deleteBlog(item: BlogResponse): Promise<void> {
    if (!this.isAuthor(item)) return;
    const confirmed = await this.confirmDialogService.confirm(
      `Da li sigurno želiš da obrišeš "${item.blog.title}"? Ova radnja se ne može poništiti.`,
      { confirmLabel: 'Obriši' }
    );
    if (!confirmed) return;

    this.blogService.deleteBlog(item.blog.id).subscribe({
      next: () => {
        this.toastService.success('Blog je obrisan.');
        if (this.selectedBlog?.blog.id === item.blog.id) {
          this.selectedBlog = null;
        }
        this.loadBlogs();
      },
      error: () => this.toastService.error('Greška pri brisanju bloga.')
    });
  }

  // ── HELPERS ───────────────────────────────────────────────────────

  /**
   * Provjerava da li string izgleda kao validan URL sa http/https protokolom.
   * Sprečava greške u konzoli kada korisnici unesu nasumičan tekst umjesto URL-ova.
   */
  isValidUrl(url: string): boolean {
    if (!url || typeof url !== 'string') return false;
    try {
      const parsed = new URL(url);
      return parsed.protocol === 'http:' || parsed.protocol === 'https:';
    } catch {
      return false;
    }
  }

  /**
   * Handler za greške pri učitavanju slika.
   * Sakriva <img> element ako slika ne može da se učita (404, CORS, itd.).
   */
  onImgError(event: Event): void {
    const img = event.target as HTMLImageElement;
    if (img) {
      // Sakrij sliku i prikaži placeholder kontejner
      img.style.display = 'none';
      // Pokušaj pronaći roditeljski wrapper i dodati fallback klasu
      const wrapper = img.closest('.card-img-wrap, .gallery-main');
      if (wrapper) {
        wrapper.classList.add('img-load-error');
      }
    }
  }

  getExcerpt(md: string): string {
    if (!md) return '';
    // Strip markdown syntax for preview
    const plain = md
      .replace(/#{1,6}\s/g, '')
      .replace(/\*\*|__|\\*|_|~~|`/g, '')
      .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
      .replace(/!\[[^\]]*\]\([^)]+\)/g, '')
      .replace(/^\s*[-*+>\d.]\s/gm, '')
      .trim();
    return plain.length > 160 ? plain.slice(0, 160) + '...' : plain;
  }

  /**
   * Escapes raw HTML in user-typed markdown before any of the regex
   * conversions below run, so a stray <script>/onerror=/etc in someone's
   * draft can't ever end up injected as real HTML via bypassSecurityTrustHtml
   * below - this used to convert raw user input straight into "trusted" HTML.
   */
  private escapeHtml(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  /**
   * Lightweight client-side Markdown → HTML renderer.
   * Backend renderuje pravi CommonMark (i sanitizuje ga), ali ovo služi za
   * preview u modalu i fallback kada nema descriptionHtml iz API-ja.
   */
  markdownToHtml(md: string): SafeHtml {
    if (!md) return this.sanitizer.bypassSecurityTrustHtml('');

    let html = this.escapeHtml(md)
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