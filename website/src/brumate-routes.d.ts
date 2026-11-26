// Written by the brumate Vite plugin from the route files. Do not edit.
export {};

declare module 'brumate/routes' {
  interface Routes {
    '/': {};
    '/guide': {};
    '/guide/[page]': { page: string };
    '/install': {};
    '/packages': {};
    '/packages/[page]': { page: string };
    '/reference': {};
    '/reference/[...page]': { page: string };
    '/spec': {};
  }
}
