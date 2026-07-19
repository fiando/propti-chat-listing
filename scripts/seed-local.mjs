#!/usr/bin/env node

import { DynamoDBClient } from '@aws-sdk/client-dynamodb';
import { DynamoDBDocumentClient, PutCommand } from '@aws-sdk/lib-dynamodb';
import { randomUUID } from 'node:crypto';

const ENDPOINT = 'http://localhost:8000';
const REGION = 'ap-southeast-1';

const client = DynamoDBDocumentClient.from(
  new DynamoDBClient({
    region: REGION,
    endpoint: ENDPOINT,
    credentials: { accessKeyId: 'local', secretAccessKey: 'local' },
  }),
  { marshallOptions: { removeUndefinedValues: true } },
);

const DEMO_USER_ID = 'demo';
const TABLE_LISTINGS = 'propti-listings-dev';
const TABLE_USERS = 'propti-users-dev';
const TABLE_LEADS = 'propti-leads-dev';

function isoNow() {
  return new Date().toISOString();
}

function futureDate(monthsFromNow = 2) {
  const d = new Date();
  d.setMonth(d.getMonth() + monthsFromNow);
  return d.toISOString();
}

function daysFromNow(days) {
  const d = new Date();
  d.setDate(d.getDate() + days);
  return d.toISOString();
}

async function put(table, item) {
  const cmd = new PutCommand({ TableName: table, Item: item });
  await client.send(cmd);
}

const listings = [
  {
    id: 'L001',
    title: 'Rumah Minimalis Modern di Jakarta Selatan',
    description: 'Rumah minimalis 2 lantai dengan desain modern, lokasi strategis dekat pusat bisnis dan sekolah internasional. Lingkungan aman dan asri.',
    price: 2500000000,
    priceUnit: 'total',
    listingType: 'sell',
    status: 'active',
    bedrooms: 4,
    bathrooms: 3,
    landArea: 200,
    buildingArea: 150,
    orientation: 'Utara',
    legalStatus: 'SHM',
    address: 'Jl. Kemang Raya No. 15',
    province: 'DKI Jakarta',
    city: 'Jakarta Selatan',
    district: 'Kemang',
    latitude: -6.2684,
    longitude: 106.8238,
  },
  {
    id: 'L002',
    title: 'Apartemen Tipe Studio di Pusat Jakarta',
    description: 'Apartemen studio fully furnished di tower premium. Fasilitas lengkap: kolam renang, gym, security 24 jam. Cocok untuk profesional muda.',
    price: 45000000,
    priceUnit: 'per_month',
    listingType: 'rent',
    status: 'active',
    bedrooms: 1,
    bathrooms: 1,
    landArea: 0,
    buildingArea: 35,
    orientation: 'Timur',
    legalStatus: 'HGB',
    address: 'Jl. Sudirman No. 1',
    province: 'DKI Jakarta',
    city: 'Jakarta Pusat',
    district: 'Tanah Abang',
    latitude: -6.2088,
    longitude: 106.8151,
  },
  {
    id: 'L003',
    title: 'Tanah Kavling Strategis di Depok',
    description: 'Tanah kavling siap bangun di kawasan berkembang. Akses mudah ke tol dan stasiun. Cocok untuk investasi atau hunian pribadi.',
    price: 350000000,
    priceUnit: 'total',
    listingType: 'sell',
    status: 'active',
    bedrooms: 0,
    bathrooms: 0,
    landArea: 300,
    buildingArea: 0,
    orientation: 'Barat',
    legalStatus: 'SHM',
    address: 'Jl. Raya Sawangan KM 5',
    province: 'Jawa Barat',
    city: 'Depok',
    district: 'Sawangan',
    latitude: -6.4120,
    longitude: 106.7640,
  },
  {
    id: 'L004',
    title: 'Rumah Mewah Dago Bandung',
    description: 'Rumah mewah bergaya klasik Eropa di kawasan elite Dago. Pemandangan kota Bandung yang indah. 5 kamar tidur, kolam renang pribadi.',
    price: 5000000000,
    priceUnit: 'total',
    listingType: 'sell',
    status: 'active',
    bedrooms: 5,
    bathrooms: 4,
    landArea: 500,
    buildingArea: 350,
    orientation: 'Selatan',
    legalStatus: 'SHM',
    address: 'Jl. Dago Pakar No. 10',
    province: 'Jawa Barat',
    city: 'Bandung',
    district: 'Coblong',
    latitude: -6.8720,
    longitude: 107.6181,
  },
  {
    id: 'L005',
    title: 'Rumah Kost Eksklusif Dekat ITS Surabaya',
    description: 'Rumah kost 10 kamar lengkap dengan AC, WiFi, dapur bersama. Cocok untuk investasi kost-kostan. Dekat kampus ITS dan UNAIR.',
    price: 35000000,
    priceUnit: 'per_year',
    listingType: 'rent',
    status: 'active',
    bedrooms: 10,
    bathrooms: 5,
    landArea: 400,
    buildingArea: 300,
    orientation: 'Timur',
    legalStatus: 'SHM',
    address: 'Jl. Keputih Tegal No. 20',
    province: 'Jawa Timur',
    city: 'Surabaya',
    district: 'Sukolilo',
    latitude: -7.2825,
    longitude: 112.7954,
  },
  {
    id: 'L006',
    title: 'Rumah Subsidi Siap Huni di Sleman',
    description: 'Rumah subsidi type 36/60 dengan spesifikasi di atas standar. Akses mudah ke ring road dan Malioboro. Bisa KPR.',
    price: 200000000,
    priceUnit: 'total',
    listingType: 'sell',
    status: 'sold',
    bedrooms: 2,
    bathrooms: 1,
    landArea: 60,
    buildingArea: 36,
    orientation: 'Utara',
    legalStatus: 'SHM',
    address: 'Jl. Kaliurang KM 12',
    province: 'DI Yogyakarta',
    city: 'Sleman',
    district: 'Ngaglik',
    latitude: -7.7300,
    longitude: 110.4050,
  },
  {
    id: 'L007',
    title: 'Ruko 3 Lantai di Pantai Indah Kapuk',
    description: 'Ruko siap pakai 3 lantai di kawasan bisnis PIK. Cocok untuk kantor, restoran, atau showroom. Parkir luas, akses tol dekat.',
    price: 4500000000,
    priceUnit: 'total',
    listingType: 'sell',
    status: 'active',
    bedrooms: 0,
    bathrooms: 2,
    landArea: 120,
    buildingArea: 300,
    orientation: 'Barat',
    legalStatus: 'HGB',
    address: 'Jl. Pantai Indah Utara No. 8',
    province: 'DKI Jakarta',
    city: 'Jakarta Utara',
    district: 'Penjaringan',
    latitude: -6.1100,
    longitude: 106.7500,
  },
  {
    id: 'L008',
    title: 'Villa Sewa Bulanan di Puncak Bogor',
    description: 'Villa 3 kamar dengan pemandangan gunung. Cocok untuk family staycation. Tersedia BBQ pit, kolam renang, dan taman luas.',
    price: 15000000,
    priceUnit: 'per_month',
    listingType: 'rent',
    status: 'active',
    bedrooms: 3,
    bathrooms: 2,
    landArea: 250,
    buildingArea: 120,
    orientation: 'Selatan',
    legalStatus: 'SHM',
    address: 'Jl. Raya Puncak KM 75',
    province: 'Jawa Barat',
    city: 'Bogor',
    district: 'Cisarua',
    latitude: -6.6770,
    longitude: 106.9491,
  },
];

const placeholderImages = [
  { imageId: 'img01', s3Key: 'demo/property1.jpg', thumbnailKey: 'demo/property1_thumb.jpg', contentType: 'image/jpeg', sizeBytes: 245000, isFeatured: true },
  { imageId: 'img02', s3Key: 'demo/property2.jpg', thumbnailKey: 'demo/property2_thumb.jpg', contentType: 'image/jpeg', sizeBytes: 198000, isFeatured: false },
  { imageId: 'img03', s3Key: 'demo/property3.jpg', thumbnailKey: 'demo/property3_thumb.jpg', contentType: 'image/jpeg', sizeBytes: 312000, isFeatured: false },
];

async function seed() {
  console.log('Seeding demo user...');

  await put(TABLE_USERS, {
    PK: DEMO_USER_ID,
    SK: 'metadata',
    userId: DEMO_USER_ID,
    googleId: 'mock-google-id',
    email: 'demo@propti.app',
    name: 'Demo User',
    profilePicture: '',
    role: 'both',
    preferences: {
      favoriteLocations: ['Jakarta Selatan', 'Jakarta Pusat'],
      searchHistory: ['rumah murah jakarta', 'apartemen sudirman'],
      notifications: true,
    },
    savedListingIds: [],
    subscription: {
      tier: 'premium',
      monthlyListingsUsed: 5,
      activeListingsCount: 4,
    },
    contactRevealThrottle: {
      windowStartedAt: isoNow(),
      revealCount: 0,
    },
    createdAt: isoNow(),
    lastLoginAt: isoNow(),
  });

  console.log('Seeding listings...');
  const savedIds = [];

  for (const l of listings) {
    const listingId = l.id;
    const now = isoNow();
    const expires = l.status === 'sold' ? undefined : futureDate(3);

    const images = placeholderImages.map((img) => ({
      ...img,
      uploadedAt: now,
    }));

    const item = {
      PK: `${DEMO_USER_ID}#${listingId}`,
      SK: listingId,
      listingId,
      userId: DEMO_USER_ID,
      title: l.title,
      description: l.description,
      price: l.price,
      priceUnit: l.priceUnit,
      listingType: l.listingType,
      status: l.status,
      propertyDetails: {
        landArea: l.landArea,
        buildingArea: l.buildingArea,
        bedrooms: l.bedrooms,
        bathrooms: l.bathrooms,
        frontWidth: l.landArea / 4,
        orientation: l.orientation,
        legalStatus: l.legalStatus,
        powerConsumption: '2200',
        amenities: l.bedrooms > 0 ? ['AC', 'WiFi', 'Parkir'] : [],
      },
      location: {
        address: l.address,
        googlePlaceId: `google-place-${listingId}`,
        latitude: l.latitude,
        longitude: l.longitude,
        province: l.province,
        city: l.city,
        district: l.district,
        nearbyPlaces: ['Sekolah', 'Supermarket', 'Tol'],
      },
      images,
      videos: [],
      imageCount: images.length,
      premiumFeatures: {
        isPremium: l.listingType === 'sell' && l.price > 1000000000,
        isFeatured: l.id === 'L001',
        featuredUntil: l.id === 'L001' ? futureDate(1) : undefined,
      },
      views: Math.floor(Math.random() * 500) + 100,
      saves: Math.floor(Math.random() * 30),
      contactReveals: Math.floor(Math.random() * 15),
      moderationStatus: 'approved',
      moderationReason: undefined,
      createdAt: now,
      updatedAt: now,
      expiresAt: expires,
    };

    await put(TABLE_LISTINGS, item);

    if (['L001', 'L002', 'L007'].includes(listingId)) {
      savedIds.push(listingId);
    }

    console.log(`  Created listing: ${l.title}`);
  }

  console.log('Seeding saved listings...');
  await put(TABLE_USERS, {
    PK: DEMO_USER_ID,
    SK: 'metadata',
    userId: DEMO_USER_ID,
    googleId: 'mock-google-id',
    email: 'demo@propti.app',
    name: 'Demo User',
    profilePicture: '',
    role: 'both',
    preferences: {
      favoriteLocations: ['Jakarta Selatan', 'Jakarta Pusat'],
      searchHistory: ['rumah murah jakarta', 'apartemen sudirman'],
      notifications: true,
    },
    savedListingIds: savedIds,
    subscription: {
      tier: 'premium',
      monthlyListingsUsed: 5,
      activeListingsCount: 4,
    },
    contactRevealThrottle: {
      windowStartedAt: isoNow(),
      revealCount: 0,
    },
    createdAt: isoNow(),
    lastLoginAt: isoNow(),
  });

  console.log('Seeding leads...');

  const lead1Id = randomUUID();
  await put(TABLE_LEADS, {
    PK: `${DEMO_USER_ID}#${lead1Id}`,
    SK: lead1Id,
    leadId: lead1Id,
    ownerUserId: DEMO_USER_ID,
    listingId: 'L001',
    name: 'Budi Santoso',
    phone: '+6281234567890',
    source: 'website',
    stage: 'interested',
    notes: ['Tertarik dengan rumah di Kemang, minta jadwal viewing'],
    activities: [
      { at: isoNow(), type: 'note_added', message: 'Lead masuk dari website' },
    ],
    followUpTasks: [
      {
        taskId: randomUUID(),
        leadId: lead1Id,
        offsetDays: 1,
        dueAt: daysFromNow(1),
        status: 'pending',
        createdAt: isoNow(),
        updatedAt: isoNow(),
      },
    ],
    createdAt: isoNow(),
    updatedAt: isoNow(),
  });
  console.log('  Created lead: Budi Santoso');

  const lead2Id = randomUUID();
  await put(TABLE_LEADS, {
    PK: `${DEMO_USER_ID}#${lead2Id}`,
    SK: lead2Id,
    leadId: lead2Id,
    ownerUserId: DEMO_USER_ID,
    listingId: 'L004',
    name: 'Siti Rahayu',
    phone: '+6281234567891',
    source: 'whatsapp',
    stage: 'new',
    notes: ['Baru menghubungi via WhatsApp, info rumah Bandung'],
    activities: [
      { at: isoNow(), type: 'note_added', message: 'Lead masuk dari WhatsApp' },
    ],
    followUpTasks: [],
    createdAt: isoNow(),
    updatedAt: isoNow(),
  });
  console.log('  Created lead: Siti Rahayu');

  const lead3Id = randomUUID();
  await put(TABLE_LEADS, {
    PK: `${DEMO_USER_ID}#${lead3Id}`,
    SK: lead3Id,
    leadId: lead3Id,
    ownerUserId: DEMO_USER_ID,
    listingId: 'L007',
    name: 'Ahmad Wijaya',
    phone: null,
    source: 'website',
    stage: 'negotiation',
    notes: ['Negosiasi harga ruko PIK. Sudah 2x viewing.'],
    activities: [
      { at: isoNow(), type: 'stage_changed', message: 'Pindah ke tahap negosiasi' },
    ],
    followUpTasks: [
      {
        taskId: randomUUID(),
        leadId: lead3Id,
        offsetDays: 3,
        dueAt: daysFromNow(3),
        status: 'pending',
        createdAt: isoNow(),
        updatedAt: isoNow(),
      },
    ],
    createdAt: isoNow(),
    updatedAt: isoNow(),
  });
  console.log('  Created lead: Ahmad Wijaya');

  console.log('Seed complete!');
  console.log(`  User:    demo@propti.app / demo`);
  console.log(`  Listings: ${listings.length}`);
  console.log(`  Saved:   ${savedIds.length}`);
  console.log(`  Leads:   3`);
}

seed().catch((err) => {
  console.error('Seed failed:', err);
  process.exit(1);
});
