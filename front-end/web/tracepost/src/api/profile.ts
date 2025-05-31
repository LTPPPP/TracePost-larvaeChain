import { getToken, decodeJWT } from '@/utils/auth';

const endpoint = process.env.NEXT_PUBLIC_API_URL;

export async function getProfile() {
  const token = getToken();

  if (!token) {
    throw new Error('No token found');
  }

  // Decode JWT để xem payload
  const decodedToken = decodeJWT(token);
  console.log('JWT Token:', token);
  console.log('Decoded JWT Payload:', decodedToken);

  const response = await fetch(`${endpoint}/users/me`, {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  if (!response.ok) {
    throw new Error('Failed to fetch profile');
  }

  return response.json();
}
