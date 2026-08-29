'use client';

import React from 'react';
import Image from 'next/image';
import Link from 'next/link';

interface ReviveLogoProps {
  variant?: 'full' | 'icon' | 'image-full';
  size?: 'sm' | 'md' | 'lg' | 'xl';
  showTagline?: boolean;
  href?: string;
  className?: string;
}

export default function ReviveLogo({
  variant = 'full',
  size = 'md',
  showTagline = false,
  href = '/',
  className = '',
}: ReviveLogoProps) {
  // Dimensions configuration
  const dimensions = {
    sm: { iconSize: 28, fontSize: 16, tagSize: 8, gap: 8 },
    md: { iconSize: 36, fontSize: 20, tagSize: 9, gap: 10 },
    lg: { iconSize: 48, fontSize: 26, tagSize: 11, gap: 12 },
    xl: { iconSize: 64, fontSize: 34, tagSize: 13, gap: 14 },
  }[size];

  // 1. Direct Image Full Logo from the uploaded brand asset
  if (variant === 'image-full') {
    const fullLogoContent = (
      <div
        className={`inline-flex flex-col items-center select-none ${className}`}
        style={{
          position: 'relative',
          maxWidth: size === 'xl' ? '320px' : size === 'lg' ? '240px' : '180px',
        }}
      >
        <Image
          src="/logo.png"
          alt="ReviveOS Logo"
          width={size === 'xl' ? 320 : size === 'lg' ? 240 : 180}
          height={size === 'xl' ? 180 : size === 'lg' ? 135 : 100}
          style={{
            objectFit: 'contain',
            height: 'auto',
            borderRadius: '12px',
          }}
          priority
        />
      </div>
    );

    if (href) {
      return (
        <Link href={href} style={{ textDecoration: 'none', display: 'inline-flex' }}>
          {fullLogoContent}
        </Link>
      );
    }
    return fullLogoContent;
  }

  // 2. Icon + Wordmark Logo (Using the exact uploaded R-Mark icon)
  const LogoContent = (
    <div
      className={`inline-flex items-center select-none ${className}`}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: `${dimensions.gap}px`,
        textDecoration: 'none',
      }}
    >
      {/* Exact Uploaded ReviveOS "R" Icon */}
      <div
        style={{
          position: 'relative',
          width: `${dimensions.iconSize}px`,
          height: `${dimensions.iconSize}px`,
          flexShrink: 0,
          borderRadius: '8px',
          overflow: 'hidden',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Image
          src="/logo-icon.png"
          alt="ReviveOS Icon"
          width={dimensions.iconSize * 2}
          height={dimensions.iconSize * 2}
          style={{
            width: '100%',
            height: '100%',
            objectFit: 'cover',
            borderRadius: '6px',
          }}
          priority
        />
      </div>

      {/* Typography Wordmark */}
      {variant !== 'icon' && (
        <div style={{ display: 'flex', flexDirection: 'column', lineHeight: 1 }}>
          <div
            style={{
              fontSize: `${dimensions.fontSize}px`,
              fontWeight: 800,
              letterSpacing: '-0.6px',
              display: 'flex',
              alignItems: 'center',
            }}
          >
            <span style={{ color: 'var(--text-primary)' }}>revive</span>
            <span
              style={{
                background: 'linear-gradient(135deg, #818cf8 0%, #c084fc 45%, #ec4899 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                fontWeight: 900,
                marginLeft: '1px',
              }}
            >
              OS
            </span>
          </div>

          {/* Optional Tagline */}
          {showTagline && (
            <span
              style={{
                fontSize: `${dimensions.tagSize}px`,
                letterSpacing: '0.22em',
                color: 'var(--text-muted)',
                textTransform: 'uppercase',
                fontWeight: 700,
                marginTop: '4px',
                fontFamily: 'var(--font-mono, monospace)',
              }}
            >
              RECOVER. OPTIMIZE. SCALE.
            </span>
          )}
        </div>
      )}
    </div>
  );

  if (href) {
    return (
      <Link href={href} style={{ textDecoration: 'none', display: 'inline-flex' }}>
        {LogoContent}
      </Link>
    );
  }

  return LogoContent;
}
