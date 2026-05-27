export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }

  get isAuthRequired() {
    return this.status === 401;
  }
  get isForbidden() {
    return this.status === 403;
  }
  get isNotFound() {
    return this.status === 404;
  }
  get isConflict() {
    return this.status === 409;
  }
  get isValidation() {
    return this.status === 400 || this.status === 422;
  }
  get isServer() {
    return this.status >= 500;
  }
}

export function userMessageFor(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.isAuthRequired) return "Please sign in to continue.";
    if (err.isForbidden) return "You don't have access to that.";
    if (err.isNotFound) return "We couldn't find that.";
    if (err.isConflict) return err.message || "That action conflicts with the current state.";
    if (err.isValidation) return err.message || "Some of the information is invalid.";
    if (err.isServer) return "Something went wrong on our side. Please try again.";
    return err.message;
  }
  return "Network error. Please check your connection.";
}
