package com.example.blog.service;

import com.example.blog.client.FollowerClient;
import com.example.blog.exception.BlogAccessDeniedException;
import com.example.blog.exception.BlogNotFoundException;
import com.example.blog.model.Blog;
import com.example.blog.model.Comment;
import com.example.blog.repository.BlogRepository;
import com.vladsch.flexmark.html.HtmlRenderer;
import com.vladsch.flexmark.parser.Parser;
import com.vladsch.flexmark.util.ast.Node;
import lombok.RequiredArgsConstructor;
import org.owasp.html.PolicyFactory;
import org.owasp.html.Sanitizers;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.Set;

@Service
@RequiredArgsConstructor
public class BlogService{

    private final BlogRepository blogRepository;
    private final FollowerClient followerClient;

    private final Parser markdownParser = Parser.builder().build();
    private final HtmlRenderer htmlRenderer = HtmlRenderer.builder().build();
    // Allowlist policy: strips <script>/event handlers/etc. from whatever a
    // user's markdown renders to, since the result is served verbatim to
    // every reader (and, via bypassSecurityTrustHtml, injected raw on the frontend).
    private final PolicyFactory htmlSanitizer = Sanitizers.FORMATTING
            .and(Sanitizers.BLOCKS)
            .and(Sanitizers.LINKS)
            .and(Sanitizers.IMAGES);

    //kreiranje bloga
    public Blog createBlog(String title, String descriptionMarkdown, List<String>imageUrls, String authorUsername){
        Blog blog = new Blog();
        blog.setTitle(title);
        blog.setDescriptionMarkdown(descriptionMarkdown);
        blog.setImageUrls(imageUrls != null ? imageUrls : List.of());
        blog.setAuthorUsername(authorUsername);
        return blogRepository.save(blog);
    }

    public List<Blog> getAllBlogsForUser(String username){
        Set<String> visibleAuthors = followerClient.getVisibleAuthors(username);
        return blogRepository.findAll().stream()
                .filter(blog -> visibleAuthors.contains(blog.getAuthorUsername()))
                .toList();
    }

    public Blog getBlogByIdForUser(String id, String username){
        Blog blog = blogRepository.findById(id)
                .orElseThrow(() -> new BlogNotFoundException("Blog nije pronadjen"));

        if (!followerClient.isFollowing(username, blog.getAuthorUsername())) {
            throw new BlogAccessDeniedException("Pristup blogu nije dozvoljen");
        }

        return blog;
    }

    //pomocna metoda - konvertuje markdown u HTML
    public String renderMarkdown(String markdown){
        if(markdown == null) return "";
        Node document = markdownParser.parse(markdown);
        String html = htmlRenderer.render(document);
        return htmlSanitizer.sanitize(html);
    }

    public Blog addComment(String blogId, String authorUsername, String text){
        Blog blog = blogRepository.findById(blogId)
                .orElseThrow(() -> new BlogNotFoundException("Blog nije pronadjen"));

        if (!followerClient.isFollowing(authorUsername, blog.getAuthorUsername())) {
            throw new BlogAccessDeniedException("Morate zapratiti autora da biste ostavili komentar");
        }

        Comment comment = new Comment();
        comment.setAuthorUsername(authorUsername);
        comment.setText(text);

        blog.getComments().add(comment);
        return blogRepository.save(blog);
    }

    public Blog editComment(String blogId, String commentId, String username, String newText){
        Blog blog = blogRepository.findById(blogId)
            .orElseThrow(() -> new BlogNotFoundException("Blog nije pronadjen"));

        Comment comment = blog.getComments().stream().filter(c -> c.getId().equals(commentId)).findFirst()
                .orElseThrow(() -> new BlogNotFoundException("Komentar nije pronadjen"));

        if(!comment.getAuthorUsername().equals(username)){
            throw new BlogAccessDeniedException("Nije tvoj komentar");
        }

        comment.setText(newText);
        comment.setLastModifiedAt(LocalDateTime.now());
        return blogRepository.save(blog);
    }

    public Blog likeBlog(String blogId, String username) {
        Blog blog = blogRepository.findById(blogId)
                .orElseThrow(() -> new BlogNotFoundException("Blog nije pronadjen"));

        if (!followerClient.isFollowing(username, blog.getAuthorUsername())) {
            throw new BlogAccessDeniedException("Morate zapratiti autora da biste lajkovali blog");
        }

        if (blog.getLikes().contains(username)) {
            blog.getLikes().remove(username);
        } else {
            blog.getLikes().add(username);
        }

        return blogRepository.save(blog);
    }
}