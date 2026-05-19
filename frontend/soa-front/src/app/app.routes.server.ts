import { RenderMode, ServerRoute } from '@angular/ssr';

export const serverRoutes: ServerRoute[] = [
  {
    path: 'tours/:id',
    renderMode: RenderMode.Client  
  },
  {
    path: 'users/search',
    renderMode: RenderMode.Client
  },
  {
    path: 'users/:username',
    renderMode: RenderMode.Client
  },
  {
    path: '**',
    renderMode: RenderMode.Prerender
  }
];