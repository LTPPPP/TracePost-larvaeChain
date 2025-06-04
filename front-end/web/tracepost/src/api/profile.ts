// import { getToken, decodeJWT, getUserInfo  } from '@/utils/auth';
import { getToken, getUserInfo } from '@/utils/auth';

const endpoint = process.env.NEXT_PUBLIC_API_URL;

export async function getProfile() {
  const token = getToken();
  const userInfo = getUserInfo();

  if (!token || !userInfo) {
    throw new Error('No authentication data found');
  }

  const userId = userInfo.user_id;

  console.log('Using token:', token);
  console.log('Fetching profile for user ID:', userId);

  const response = await fetch(`${endpoint}/users/${userId}`, {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || `Failed to fetch profile: ${response.status}`);
  }

  return response.json();
}

// export async function getProfile() {
//   const token = getToken();

//   if (!token) {
//     throw new Error('No token found');
//   }

//   // Decode JWT để xem payload
//   const decodedToken = decodeJWT(token);
//   console.log('JWT Token:', token);
//   console.log('Decoded JWT Payload:', decodedToken);

//   const response = await fetch(`${endpoint}/users/me`, {
//     method: 'GET',
//     headers: {
//       Authorization: `Bearer ${token}`,
//       'Content-Type': 'application/json'
//     }
//   });

//   if (!response.ok) {
//     throw new Error('Failed to fetch profile');
//   }

//   return response.json();
// }
