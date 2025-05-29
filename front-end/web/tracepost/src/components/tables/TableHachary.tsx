'use client';
import React from 'react';
import { useState, useEffect } from 'react';
import { Table, TableBody, TableCell, TableHeader, TableRow } from '../ui/table';
import { EyeIcon } from '../../icons/index';
import Image from 'next/image';
import ToggleButton from '../ui/toggle/index';
import { useRouter } from 'next/navigation';

interface Order {
  id: number;
  user: {
    image: string;
    name: string;
    role: string;
  };
  contact: string;
  details: React.ReactNode;
  company: string;
  email: string;
  is_active: boolean;
}

export default function TableHatchary() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();

  useEffect(() => {
    const fetchHatcheries = async () => {
      try {
        setLoading(true);
        setError(null);

        // Retrieve token from localStorage or sessionStorage
        const token = localStorage.getItem('token') || sessionStorage.getItem('token');

        if (!token) {
          setError('Authentication token is missing. Please log in.');
          router.push('/login'); // Redirect to login page
          return;
        }

        const response = await fetch('http://localhost:8080/api/v1/hatcheries', {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}` // Include the token for authentication
          }
        });

        if (!response.ok) {
          if (response.status === 401) {
            setError('Unauthorized: Invalid or expired token. Please log in again.');
            localStorage.removeItem('token'); // Clear invalid token
            sessionStorage.removeItem('token');
            router.push('/login'); // Redirect to login page
            return;
          }
          throw new Error(`Failed to fetch hatcheries: ${response.statusText}`);
        }

        const data = await response.json();
        console.log('Raw API response:', data); // Debug the response structure

        // Define the expected hatchery data structure
        interface HatcheryData {
          id?: number;
          name?: string;
          type?: string;
          location?: string;
          company?: {
            name?: string;
            contact_info?: string; // Adjust based on actual API field name
          };
          created_at?: string;
          updated_at?: string;
          is_active?: boolean;
        }

        // Handle different response structures
        let hatcheryArray: HatcheryData[] = [];
        if (Array.isArray(data)) {
          hatcheryArray = data;
        } else if (data.data && Array.isArray(data.data)) {
          hatcheryArray = data.data;
        } else if (data.hatcheries && Array.isArray(data.hatcheries)) {
          hatcheryArray = data.hatcheries;
        } else if (data && typeof data === 'object') {
          hatcheryArray = [data]; // Handle single object response
        } else {
          throw new Error('Unexpected API response format');
        }

        // Map the API data to the Order interface
        const formattedData: Order[] = hatcheryArray.map((item) => ({
          id: item.id || 0,
          user: {
            image: '/img/default-avatar.png', // Default image if not provided
            name: item.name || 'Unknown',
            role: ''
          },
          details: <EyeIcon />,
          contact: item.company?.contact_info || 'N/A', // Fallback for missing contact_info
          company: item.company?.name || item.type || item.location || 'N/A', // Use company.name if available

          email: 'N/A', // Assuming contact is used as email
          is_active: item.is_active ?? false
        }));

        // Log the formatted data
        console.log(
          'Formatted Orders:',
          formattedData.map((order) => ({
            id: order.id,
            user: {
              image: order.user.image,
              name: order.user.name,
              role: order.user.role
            },
            details: 'EyeIcon',
            contact: order.contact,
            company: order.company,
            email: order.email,
            is_active: order.is_active
          }))
        );

        setOrders(formattedData);
      } catch (err) {
        console.error('Error fetching hatcheries:', err);
        setError(err instanceof Error ? err.message : 'An unexpected error occurred');
      } finally {
        setLoading(false);
      }
    };

    fetchHatcheries();
  }, [router]); // Include router in dependencies for redirect

  // Handle toggle for blocking/unblocking user (updates is_active)
  const handleToggle = async (id: number, checked: boolean) => {
    try {
      // Optimistic update: Update local state first
      const previousOrders = orders; // Store previous state for rollback
      const updated = orders.map((order) => (order.id === id ? { ...order, is_active: checked } : order));
      setOrders(updated);

      // Get token for authentication
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');

      if (!token) {
        setError('Authentication token is missing. Please log in.');
        setOrders(previousOrders); // Revert state
        router.push('/login');
        return;
      }

      // Send PATCH request to update is_active status
      const response = await fetch(`http://localhost:8080/api/v1/admin/users/${id}/status`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({ is_active: checked })
      });

      if (!response.ok) {
        if (response.status === 401) {
          setError('Unauthorized: Invalid or expired token. Please log in again.');
          localStorage.removeItem('token');
          sessionStorage.removeItem('token');
          router.push('/login');
          setOrders(previousOrders); // Revert state
          return;
        }
        throw new Error(`Failed to update user status: ${response.statusText}`);
      }

      // If successful, the local state is already updated
      console.log(`User ID ${id} ${checked ? 'unblocked' : 'blocked'} successfully`);
    } catch (err) {
      console.error('Error updating user status:', err);
      setError(err instanceof Error ? err.message : 'Failed to update status');
      // Revert to previous state on error
      setOrders(orders.map((order) => (order.id === id ? { ...order, is_active: !checked } : order)));
    }
  };

  if (loading) {
    return <div className='text-center py-4 text-gray-500 dark:text-gray-400'>Loading distributors...</div>;
  }

  if (error) {
    return <div className='text-center py-4 text-red-500 dark:text-red-400'>Error: {error}</div>;
  }

  return (
    <div className='overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-white/[0.05] dark:bg-white/[0.03]'>
      <div className='max-w-full overflow-x-auto'>
        <div className='min-w-[1102px]'>
          <Table>
            <TableHeader className='border-b border-gray-100 dark:border-white/[0.05]'>
              <TableRow>
                <TableCell
                  isHeader
                  className='px-5 py-3 font-medium text-gray-500 text-start text-theme-xs dark:text-gray-400'
                >
                  User
                </TableCell>
                <TableCell
                  isHeader
                  className='px-5 py-3 font-medium text-gray-500 text-start text-theme-xs dark:text-gray-400'
                >
                  Detail
                </TableCell>
                <TableCell
                  isHeader
                  className='px-5 py-3 font-medium text-gray-500 text-start text-theme-xs dark:text-gray-400'
                >
                  Company
                </TableCell>
                <TableCell
                  isHeader
                  className='px-5 py-3 font-medium text-gray-500 text-start text-theme-xs dark:text-gray-400'
                >
                  Contact
                </TableCell>
                <TableCell
                  isHeader
                  className='px-5 py-3 font-medium text-gray-500 text-start text-theme-xs dark:text-gray-400'
                >
                  Status
                </TableCell>
              </TableRow>
            </TableHeader>

            <TableBody className='divide-y divide-gray-100 dark:divide-white/[0.05]'>
              {orders.length === 0 ? (
                <TableRow>
                  <td colSpan={5} className='px-5 py-4 text-center text-gray-500 dark:text-gray-400'>
                    No distributors found.
                  </td>
                </TableRow>
              ) : (
                orders.map((order) => (
                  <TableRow key={order.id}>
                    <TableCell className='px-5 py-4 sm:px-6 text-start'>
                      <div className='flex items-center gap-3'>
                        <div className='w-10 h-10 overflow-hidden rounded-full'>
                          <Image width={40} height={40} src={order.user.image} alt={order.user.name} />
                        </div>
                        <div>
                          <span className='block font-medium text-gray-800 text-theme-sm dark:text-white/90'>
                            {order.user.name}
                          </span>
                          <span className='block text-gray-500 text-theme-xs dark:text-gray-400'>
                            {order.user.role}
                          </span>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell className='px-4 py-3 text-gray-500 text-start text-theme-sm dark:text-gray-400'>
                      {order.details}
                    </TableCell>
                    <TableCell className='px-4 py-3 text-gray-500 text-start text-theme-sm dark:text-gray-400'>
                      {order.company}
                    </TableCell>
                    <TableCell className='px-4 py-3 text-gray-500 text-start text-theme-sm dark:text-gray-400'>
                      {order.contact}
                    </TableCell>
                    <TableCell className='px-4 py-3 text-gray-500 text-theme-sm dark:text-gray-400'>
                      <ToggleButton
                        checked={order.is_active} // Use is_active instead of status
                        onChange={(checked) => handleToggle(order.id, checked)}
                      />
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
