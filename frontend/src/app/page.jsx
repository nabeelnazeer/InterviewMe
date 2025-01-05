'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { redirect } from 'next/navigation';

export default function Home() {
  redirect('/cvScoring');
}
