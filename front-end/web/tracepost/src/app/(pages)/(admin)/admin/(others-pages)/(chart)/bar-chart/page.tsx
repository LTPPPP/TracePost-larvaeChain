import BarChartOne from '@/components/charts/bar/BarChartOne';
import ComponentCard from '@/components/common/ComponentCard';
import PageBreadcrumb from '@/components/common/PageBreadCrumb';
import { Metadata } from 'next';
import React from 'react';

export const metadata: Metadata = {
  title: 'TracePost',
  description: 'A high-performance backend system for shrimp larvae traceability using blockchain technology.'
};

export default function page() {
  return (
    <div>
      <PageBreadcrumb pageTitle='Bar Chart' />
      <div className='space-y-6'>
        <ComponentCard title='Bar Chart 1'>
          <BarChartOne />
        </ComponentCard>
      </div>
    </div>
  );
}
