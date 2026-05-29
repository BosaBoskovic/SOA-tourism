package com.example.blog.rpc;

import com.example.blog.model.Blog;
import com.example.blog.service.BlogService;
import lombok.RequiredArgsConstructor;
import blogs.v1.BlogsServiceGrpc;
import blogs.v1.GetAllBlogsRequest;
import blogs.v1.GetAllBlogsResponse;
import blogs.v1.GetBlogRequest;
import blogs.v1.GetBlogResponse;
import io.grpc.stub.StreamObserver;
import net.devh.boot.grpc.server.service.GrpcService;

import java.util.List;

@GrpcService
@RequiredArgsConstructor
public class BlogGrpcService extends BlogsServiceGrpc.BlogsServiceImplBase {

    private final BlogService blogService;

    @Override
    public void getBlog(GetBlogRequest request,
                        StreamObserver<GetBlogResponse> responseObserver) {

        System.out.println("RPC GetBlog pozvan");


        Blog blog = blogService.getBlogByIdForUser(
                request.getBlogId(),
                request.getUsername()
        );

        GetBlogResponse response = GetBlogResponse.newBuilder()
                .setBlog(toGrpcBlog(blog, request.getUsername(), true))
                .build();

        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void getAllBlogs(GetAllBlogsRequest request,
                            StreamObserver<GetAllBlogsResponse> responseObserver) {

        System.out.println("RPC GetAllBlogs pozvan");

        List<Blog> blogs = blogService.getAllBlogsForUser(request.getUsername());

        GetAllBlogsResponse response = GetAllBlogsResponse.newBuilder()
                .addAllBlogs(
                        blogs.stream()
                                .map(blog -> toGrpcBlog(blog, request.getUsername(), false))
                                .toList()
                )
                .build();

        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    private blogs.v1.Blog toGrpcBlog(Blog blog, String username, boolean includeHtml) {
        blogs.v1.Blog.Builder builder = blogs.v1.Blog.newBuilder()
                .setId(blog.getId())
                .setTitle(blog.getTitle())
                .setDescriptionMarkdown(blog.getDescriptionMarkdown())
                .setAuthorUsername(blog.getAuthorUsername())
                .setCreatedAt(blog.getCreatedAt().toString())
                .addAllImageUrls(blog.getImageUrls())
                .setLikesCount(blog.getLikes().size())
                .setLikedByCurrentUser(blog.getLikes().contains(username));

        if (includeHtml) {
            builder.setDescriptionHtml(
                    blogService.renderMarkdown(blog.getDescriptionMarkdown())
            );
        }

        return builder.build();
    }
}
