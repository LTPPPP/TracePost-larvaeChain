export const decodeJWT = (token: string) => {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;

    const payload = parts[1];
    const paddedPayload = payload + '='.repeat((4 - (payload.length % 4)) % 4);
    const decodedPayload = atob(paddedPayload);

    return JSON.parse(decodedPayload);
  } catch (error) {
    console.error('Error decoding JWT:', error);
    return null;
  }
};

// Lưu token và user info
export const saveAuthData = (data: { access_token: string; expires_in: number; role: string; user_id: number }) => {
  const expiryTime = new Date().getTime() + data.expires_in * 1000;

  localStorage.setItem('access_token', data.access_token);
  localStorage.setItem('user_id', data.user_id.toString());
  localStorage.setItem('role', data.role);
  localStorage.setItem('token_expiry', expiryTime.toString());
};

// Lấy token
export const getToken = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('access_token');
};

// Lấy user info
export const getUserInfo = () => {
  if (typeof window === 'undefined') return null;

  const token = localStorage.getItem('access_token');
  const userId = localStorage.getItem('user_id');
  const role = localStorage.getItem('role');
  const tokenExpiry = localStorage.getItem('token_expiry');

  if (!token || !userId || !role || !tokenExpiry) return null;

  // Check if token expired
  const now = new Date().getTime();
  const expiryTime = parseInt(tokenExpiry);

  if (now >= expiryTime) {
    clearAuthData();
    return null;
  }

  return {
    token,
    user_id: parseInt(userId),
    role
  };
};

// Xóa auth data
export const clearAuthData = () => {
  if (typeof window === 'undefined') return;

  localStorage.removeItem('access_token');
  localStorage.removeItem('user_id');
  localStorage.removeItem('role');
  localStorage.removeItem('token_expiry');
};

// Check if user is logged in
export const isLoggedIn = (): boolean => {
  return getUserInfo() !== null;
};
