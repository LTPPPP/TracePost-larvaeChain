'use client';
import React from 'react';
import { useEffect, useState } from 'react';

interface AnalyticsData {
  success: boolean;
  message: string;
  data: {
    batch: {
      total_batches_produced: number;
      active_batches: number;
      batches_by_status: Record<string, number>;
      batches_by_region: Record<string, number>;
      batches_by_species: Record<string, number>;
      batches_by_hatchery: Record<string, number>;
      production_trend: {
        last_12_months: number[];
        last_6_months: number[];
      };
      average_shipment_time: Record<string, number>;
      last_updated: string;
    };
    blockchain: {
      total_nodes: number;
      active_nodes: number;
      network_health: string;
      consensus_status: string;
      average_block_time_ms: number;
      transactions_per_second: number;
      pending_transactions: number;
      node_latencies: Record<string, number>;
      chain_health: Record<string, string>;
      cross_chain_transactions: Record<string, number>;
      last_updated: string;
    };
    compliance: {
      total_certificates: number;
      valid_certificates: number;
      expired_certificates: number;
      revoked_certificates: number;
      company_compliance: Record<string, number>;
      standards_compliance: Record<string, number>;
      regional_compliance: Record<string, number>;
      compliance_trends: {
        last_6_months: number[];
      };
      last_updated: string;
    };
    system: {
      active_users: number;
      total_batches: number;
      blockchain_tx_count: number;
      api_requests_per_hour: number;
      avg_response_time_ms: number;
      system_health: string;
      server_cpu_usage: number;
      server_memory_usage: number;
      db_connections: number;
      last_updated: string;
    };
    user_activity: {
      active_users_by_role: Record<string, number>;
      login_frequency: {
        last_30_days: number;
        last_7_days: number;
        today: number;
        yesterday: number;
      };
      api_endpoint_usage: Record<string, number>;
      most_active_users: Array<{
        user_id: number;
        username: string;
        request_count: number;
        last_active: string;
      }>;
      user_growth: {
        last_30_days: number;
        last_7_days: number;
        last_90_days: number;
        today: number;
        yesterday: number;
      };
      last_updated: string;
    };
    timestamp: string;
  };
}
export const EcommerceMetrics = () => {
  const [analytics, setAnalytics] = useState<AnalyticsData['data'] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchAnalytics = async () => {
      try {
        setLoading(true);
        setError(null);

        // Retrieve token from localStorage or sessionStorage
        const token = localStorage.getItem('token') || sessionStorage.getItem('token');

        if (!token) {
          setError('Authentication token is missing. Please log in.');
          return;
        }

        const response = await fetch('http://localhost:8080/api/v1/admin/analytics/dashboard', {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`
          }
        });

        if (!response.ok) {
          if (response.status === 401) {
            setError('Unauthorized: Invalid or expired token. Please log in again.');
            localStorage.removeItem('token');
            sessionStorage.removeItem('token');
            return;
          }
          throw new Error(`Failed to fetch analytics: ${response.statusText}`);
        }

        const data: AnalyticsData = await response.json();
        if (!data.success) {
          throw new Error(data.message || 'Failed to retrieve analytics data');
        }

        setAnalytics(data.data);
      } catch (err) {
        console.error('Error fetching analytics:', err);
        setError(err instanceof Error ? err.message : 'An unexpected error occurred');
      } finally {
        setLoading(false);
      }
    };

    fetchAnalytics();
  }, []);

  if (loading) {
    return <div className='text-center py-4 text-gray-500 dark:text-gray-400'>Loading dashboard...</div>;
  }

  if (error) {
    return <div className='text-center py-4 text-red-500 dark:text-red-400'>Error: {error}</div>;
  }

  if (!analytics) {
    return <div className='text-center py-4 text-gray-500 dark:text-gray-400'>No analytics data available.</div>;
  }

  // Calculate derived metrics
  const totalActiveUsers = Object.values(analytics.user_activity.active_users_by_role).reduce(
    (sum, count) => sum + count,
    0
  );
  const totalBatchesByRegion = Object.values(analytics.batch.batches_by_region).reduce((sum, count) => sum + count, 0);
  const averageRegionalCompliance =
    Object.values(analytics.compliance.regional_compliance).reduce((sum, value) => sum + value, 0) /
    Object.keys(analytics.compliance.regional_compliance).length;

  return (
    <div className=''>
      <h2 className='text-2xl font-bold text-gray-800 dark:text-white/90 mb-6'>Admin Dashboard</h2>
      <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-4'>
        {/* Total Active Users */}
        <div className='rounded-2xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03] md:p-6'>
          <div className='flex items-center justify-center w-12 h-12 bg-gray-100 rounded-xl dark:bg-gray-800'>
            {/* Replace with <GroupIcon className='text-gray-800 size-6 dark:text-white/90' /> */}
            <span className='text-gray-800 dark:text-white/90'>👤</span>
          </div>
          <div className='flex items-end justify-between mt-5'>
            <div>
              <span className='text-sm text-gray-500 dark:text-gray-400'>Total Active Users</span>
              <h4 className='mt-2 font-bold text-gray-800 text-title-sm dark:text-white/90'>
                {totalActiveUsers.toLocaleString()}
              </h4>
            </div>
          </div>
        </div>

        {/* Logins Today */}
        <div className='rounded-2xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03] md:p-6'>
          <div className='flex items-center justify-center w-12 h-12 bg-gray-100 rounded-xl dark:bg-gray-800'>
            {/* Replace with <LoginIcon className='text-gray-800 size-6 dark:text-white/90' /> */}
            <span className='text-gray-800 dark:text-white/90'>🔒</span>
          </div>
          <div className='flex items-end justify-between mt-5'>
            <div>
              <span className='text-sm text-gray-500 dark:text-gray-400'>Logins Today</span>
              <h4 className='mt-2 font-bold text-gray-800 text-title-sm dark:text-white/90'>
                {analytics.user_activity.login_frequency.today.toLocaleString()}
              </h4>
            </div>
          </div>
        </div>

        {/* User Growth Today */}
        <div className='rounded-2xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03] md:p-6'>
          <div className='flex items-center justify-center w-12 h-12 bg-gray-100 rounded-xl dark:bg-gray-800'>
            {/* Replace with <UserPlusIcon className='text-gray-800 size-6 dark:text-white/90' /> */}
            <span className='text-gray-800 dark:text-white/90'>📈</span>
          </div>
          <div className='flex items-end justify-between mt-5'>
            <div>
              <span className='text-sm text-gray-500 dark:text-gray-400'>User Growth Today</span>
              <h4 className='mt-2 font-bold text-gray-800 text-title-sm dark:text-white/90'>
                {analytics.user_activity.user_growth.today.toLocaleString()}
              </h4>
            </div>
          </div>
        </div>

        {/* Total Batches by Region */}
        <div className='rounded-2xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03] md:p-6'>
          <div className='flex items-center justify-center w-12 h-12 bg-gray-100 rounded-xl dark:bg-gray-800'>
            {/* Replace with <BoxIconLine className='text-gray-800 size-6 dark:text-white/90' /> */}
            <span className='text-gray-800 dark:text-white/90'>📦</span>
          </div>
          <div className='flex items-end justify-between mt-5'>
            <div>
              <span className='text-sm text-gray-500 dark:text-gray-400'>Total Batches (Regions)</span>
              <h4 className='mt-2 font-bold text-gray-800 text-title-sm dark:text-white/90'>
                {totalBatchesByRegion.toLocaleString()}
              </h4>
            </div>
          </div>
        </div>

        {/* Average Shipment Time (North to South) */}
        <div className='rounded-2xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03] md:p-6'>
          <div className='flex items-center justify-center w-12 h-12 bg-gray-100 rounded-xl dark:bg-gray-800'>
            {/* Replace with <ClockIcon className='text-gray-800 size-6 dark:text-white/90' /> */}
            <span className='text-gray-800 dark:text-white/90'>⏳</span>
          </div>
          <div className='flex items-end justify-between mt-5'>
            <div>
              <span className='text-sm text-gray-500 dark:text-gray-400'>Avg. Shipment (North to South)</span>
              <h4 className='mt-2 font-bold text-gray-800 text-title-sm dark:text-white/90'>
                {analytics.batch.average_shipment_time['North to South']} days
              </h4>
            </div>
          </div>
        </div>

        {/* Average Regional Compliance */}
        <div className='rounded-2xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03] md:p-6'>
          <div className='flex items-center justify-center w-12 h-12 bg-gray-100 rounded-xl dark:bg-gray-800'>
            {/* Replace with <CheckIcon className='text-gray-800 size-6 dark:text-white/90' /> */}
            <span className='text-gray-800 dark:text-white/90'>✅</span>
          </div>
          <div className='flex items-end justify-between mt-5'>
            <div>
              <span className='text-sm text-gray-500 dark:text-gray-400'>Avg. Regional Compliance</span>
              <h4 className='mt-2 font-bold text-gray-800 text-title-sm dark:text-white/90'>
                {averageRegionalCompliance.toFixed(1)}%
              </h4>
            </div>
          </div>
        </div>

        {/* Server CPU Usage */}
        <div className='rounded-2xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03] md:p-6'>
          <div className='flex items-center justify-center w-12 h-12 bg-gray-100 rounded-xl dark:bg-gray-800'>
            {/* Replace with <CpuIcon className='text-gray-800 size-6 dark:text-white/90' /> */}
            <span className='text-gray-800 dark:text-white/90'>💻</span>
          </div>
          <div className='flex items-end justify-between mt-5'>
            <div>
              <span className='text-sm text-gray-500 dark:text-gray-400'>Server CPU Usage</span>
              <h4 className='mt-2 font-bold text-gray-800 text-title-sm dark:text-white/90'>
                {analytics.system.server_cpu_usage}%
              </h4>
            </div>
          </div>
        </div>

        {/* Blockchain Network Health */}
        <div className='rounded-2xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03] md:p-6'>
          <div className='flex items-center justify-center w-12 h-12 bg-gray-100 rounded-xl dark:bg-gray-800'>
            {/* Replace with <NetworkIcon className='text-gray-800 size-6 dark:text-white/90' /> */}
            <span className='text-gray-800 dark:text-white/90'>🌐</span>
          </div>
          <div className='flex items-end justify-between mt-5'>
            <div>
              <span className='text-sm text-gray-500 dark:text-gray-400'>Blockchain Network Health</span>
              <h4 className='mt-2 font-bold text-gray-800 text-title-sm dark:text-white/90 capitalize'>
                {analytics.blockchain.network_health}
              </h4>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
