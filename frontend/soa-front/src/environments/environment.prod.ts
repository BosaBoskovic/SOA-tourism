// This project doesn't have a defined production deployment target yet (no
// frontend Dockerfile/hosting) - override apiUrl for wherever this actually
// gets deployed, same as you would with any other Angular environment file.
export const environment = {
  production: true,
  apiUrl: 'http://localhost:8080',
};
