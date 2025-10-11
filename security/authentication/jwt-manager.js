/**
 * JWT Security Manager for Taishang Laojun AI Platform
 * 
 * This module provides comprehensive JWT token management including:
 * - Secure token generation and validation
 * - Token rotation and refresh mechanisms
 * - Blacklist management for revoked tokens
 * - Regional compliance and data residency
 * - Rate limiting and abuse prevention
 * - Audit logging and monitoring
 * - Multi-tenant support
 * - Advanced security features (JWE, key rotation, etc.)
 */

import jwt from 'jsonwebtoken';
import crypto from 'crypto';
import { promisify } from 'util';

class JWTManager {
    constructor(options = {}) {
        this.config = {
            // Token settings
            accessTokenExpiry: options.accessTokenExpiry || '15m',
            refreshTokenExpiry: options.refreshTokenExpiry || '7d',
            issuer: options.issuer || 'taishanglaojun.ai',
            audience: options.audience || 'taishanglaojun-users',
            
            // Security settings
            algorithm: options.algorithm || 'RS256',
            keyRotationInterval: options.keyRotationInterval || 86400000, // 24 hours
            maxTokensPerUser: options.maxTokensPerUser || 10,
            enableTokenBlacklist: options.enableTokenBlacklist !== false,
            enableRefreshTokenRotation: options.enableRefreshTokenRotation !== false,
            
            // Rate limiting
            maxTokenRequests: options.maxTokenRequests || 100, // per hour
            maxRefreshRequests: options.maxRefreshRequests || 20, // per hour
            
            // Regional compliance
            region: options.region || 'us-east-1',
            enableGDPRCompliance: options.enableGDPRCompliance !== false,
            enableCCPACompliance: options.enableCCPACompliance !== false,
            dataResidency: options.dataResidency || {},
            
            // Encryption (JWE)
            enableEncryption: options.enableEncryption || false,
            encryptionAlgorithm: options.encryptionAlgorithm || 'A256GCM',
            keyManagementAlgorithm: options.keyManagementAlgorithm || 'RSA-OAEP-256',
            
            // Audit and monitoring
            enableAuditLogging: options.enableAuditLogging !== false,
            enableMetrics: options.enableMetrics !== false,
            
            // Multi-tenant
            enableMultiTenant: options.enableMultiTenant || false,
            
            ...options
        };

        // Key management
        this.keyPairs = new Map();
        this.currentKeyId = null;
        this.encryptionKeys = new Map();
        
        // Token tracking
        this.activeTokens = new Map(); // userId -> Set of token IDs
        this.tokenBlacklist = new Set();
        this.refreshTokens = new Map(); // tokenId -> refresh token data
        
        // Rate limiting
        this.rateLimitTracking = new Map();
        
        // Audit logging
        this.auditLogs = [];
        
        // Metrics
        this.metrics = {
            tokensIssued: 0,
            tokensValidated: 0,
            tokensRevoked: 0,
            refreshTokensUsed: 0,
            validationErrors: 0,
            securityViolations: 0
        };

        this.init();
    }

    /**
     * Initialize JWT manager
     */
    async init() {
        try {
            // Generate initial key pair
            await this.generateKeyPair();

            // Setup key rotation
            this.setupKeyRotation();

            // Setup cleanup intervals
            this.setupCleanupIntervals();

            // Initialize encryption keys if enabled
            if (this.config.enableEncryption) {
                await this.generateEncryptionKey();
            }

            console.log('🔐 JWT Manager initialized');

        } catch (error) {
            console.error('❌ Failed to initialize JWT Manager:', error);
            throw error;
        }
    }

    /**
     * Generate RSA key pair for JWT signing
     */
    async generateKeyPair() {
        const keyId = this.generateKeyId();
        
        const { publicKey, privateKey } = crypto.generateKeyPairSync('rsa', {
            modulusLength: 2048,
            publicKeyEncoding: {
                type: 'spki',
                format: 'pem'
            },
            privateKeyEncoding: {
                type: 'pkcs8',
                format: 'pem'
            }
        });

        this.keyPairs.set(keyId, {
            publicKey,
            privateKey,
            createdAt: Date.now(),
            algorithm: this.config.algorithm
        });

        // Set as current key if first key or if current key is old
        if (!this.currentKeyId || this.shouldRotateKey()) {
            this.currentKeyId = keyId;
        }

        this.log('Key pair generated', { keyId, algorithm: this.config.algorithm });

        return keyId;
    }

    /**
     * Generate encryption key for JWE
     */
    async generateEncryptionKey() {
        const keyId = this.generateKeyId();
        const key = crypto.randomBytes(32); // 256-bit key

        this.encryptionKeys.set(keyId, {
            key,
            createdAt: Date.now(),
            algorithm: this.config.encryptionAlgorithm
        });

        this.log('Encryption key generated', { keyId });

        return keyId;
    }

    /**
     * Issue access token
     */
    async issueAccessToken(payload, options = {}) {
        try {
            // Validate payload
            this.validateTokenPayload(payload);

            // Check rate limits
            if (!this.checkRateLimit(payload.sub, 'access')) {
                throw new Error('Rate limit exceeded for token requests');
            }

            // Generate token ID
            const tokenId = this.generateTokenId();
            
            // Prepare JWT payload
            const jwtPayload = {
                ...payload,
                jti: tokenId,
                iss: this.config.issuer,
                aud: this.config.audience,
                iat: Math.floor(Date.now() / 1000),
                exp: this.calculateExpiry(this.config.accessTokenExpiry),
                type: 'access',
                region: this.config.region,
                tenant: options.tenant || 'default'
            };

            // Add compliance flags
            if (this.config.enableGDPRCompliance) {
                jwtPayload.gdpr = true;
            }
            if (this.config.enableCCPACompliance) {
                jwtPayload.ccpa = true;
            }

            // Sign token
            const keyPair = this.keyPairs.get(this.currentKeyId);
            const token = jwt.sign(jwtPayload, keyPair.privateKey, {
                algorithm: this.config.algorithm,
                keyid: this.currentKeyId
            });

            // Encrypt token if enabled
            const finalToken = this.config.enableEncryption ? 
                await this.encryptToken(token) : token;

            // Track active token
            this.trackActiveToken(payload.sub, tokenId);

            // Update metrics
            this.metrics.tokensIssued++;

            // Update rate limit
            this.updateRateLimit(payload.sub, 'access');

            // Audit log
            this.log('Access token issued', {
                userId: payload.sub,
                tokenId,
                tenant: options.tenant,
                region: this.config.region
            });

            return {
                token: finalToken,
                tokenId,
                expiresIn: this.config.accessTokenExpiry,
                tokenType: 'Bearer'
            };

        } catch (error) {
            this.metrics.validationErrors++;
            this.log('Access token issuance failed', { 
                userId: payload.sub, 
                error: error.message 
            }, 'error');
            throw error;
        }
    }

    /**
     * Issue refresh token
     */
    async issueRefreshToken(payload, options = {}) {
        try {
            // Validate payload
            this.validateTokenPayload(payload);

            // Check rate limits
            if (!this.checkRateLimit(payload.sub, 'refresh')) {
                throw new Error('Rate limit exceeded for refresh token requests');
            }

            // Generate token ID
            const tokenId = this.generateTokenId();
            
            // Prepare JWT payload
            const jwtPayload = {
                sub: payload.sub,
                jti: tokenId,
                iss: this.config.issuer,
                aud: this.config.audience,
                iat: Math.floor(Date.now() / 1000),
                exp: this.calculateExpiry(this.config.refreshTokenExpiry),
                type: 'refresh',
                region: this.config.region,
                tenant: options.tenant || 'default'
            };

            // Sign token
            const keyPair = this.keyPairs.get(this.currentKeyId);
            const token = jwt.sign(jwtPayload, keyPair.privateKey, {
                algorithm: this.config.algorithm,
                keyid: this.currentKeyId
            });

            // Encrypt token if enabled
            const finalToken = this.config.enableEncryption ? 
                await this.encryptToken(token) : token;

            // Store refresh token data
            this.refreshTokens.set(tokenId, {
                userId: payload.sub,
                token: finalToken,
                createdAt: Date.now(),
                expiresAt: this.calculateExpiry(this.config.refreshTokenExpiry) * 1000,
                used: false,
                tenant: options.tenant || 'default'
            });

            // Update rate limit
            this.updateRateLimit(payload.sub, 'refresh');

            // Audit log
            this.log('Refresh token issued', {
                userId: payload.sub,
                tokenId,
                tenant: options.tenant
            });

            return {
                token: finalToken,
                tokenId,
                expiresIn: this.config.refreshTokenExpiry
            };

        } catch (error) {
            this.metrics.validationErrors++;
            this.log('Refresh token issuance failed', { 
                userId: payload.sub, 
                error: error.message 
            }, 'error');
            throw error;
        }
    }

    /**
     * Validate token
     */
    async validateToken(token, options = {}) {
        try {
            // Decrypt token if encrypted
            const decryptedToken = this.config.enableEncryption ? 
                await this.decryptToken(token) : token;

            // Decode token header to get key ID
            const decoded = jwt.decode(decryptedToken, { complete: true });
            
            if (!decoded || !decoded.header.kid) {
                throw new Error('Invalid token format');
            }

            const keyId = decoded.header.kid;
            const keyPair = this.keyPairs.get(keyId);
            
            if (!keyPair) {
                throw new Error('Unknown signing key');
            }

            // Verify token
            const payload = jwt.verify(decryptedToken, keyPair.publicKey, {
                algorithms: [this.config.algorithm],
                issuer: this.config.issuer,
                audience: this.config.audience
            });

            // Check if token is blacklisted
            if (this.config.enableTokenBlacklist && this.tokenBlacklist.has(payload.jti)) {
                throw new Error('Token has been revoked');
            }

            // Check token type
            if (options.expectedType && payload.type !== options.expectedType) {
                throw new Error(`Expected ${options.expectedType} token, got ${payload.type}`);
            }

            // Check tenant if multi-tenant is enabled
            if (this.config.enableMultiTenant && options.tenant && payload.tenant !== options.tenant) {
                throw new Error('Token tenant mismatch');
            }

            // Check region compliance
            if (options.region && payload.region !== options.region) {
                this.log('Cross-region token usage', {
                    tokenRegion: payload.region,
                    requestRegion: options.region,
                    userId: payload.sub
                }, 'warning');
            }

            // Update metrics
            this.metrics.tokensValidated++;

            // Audit log
            this.log('Token validated', {
                userId: payload.sub,
                tokenId: payload.jti,
                type: payload.type
            });

            return payload;

        } catch (error) {
            this.metrics.validationErrors++;
            this.log('Token validation failed', { 
                error: error.message 
            }, 'error');
            throw error;
        }
    }

    /**
     * Refresh access token using refresh token
     */
    async refreshAccessToken(refreshToken, options = {}) {
        try {
            // Validate refresh token
            const refreshPayload = await this.validateToken(refreshToken, { 
                expectedType: 'refresh',
                ...options 
            });

            // Check if refresh token exists and is not used
            const refreshData = this.refreshTokens.get(refreshPayload.jti);
            
            if (!refreshData || refreshData.used) {
                throw new Error('Invalid or already used refresh token');
            }

            // Check if refresh token is expired
            if (Date.now() > refreshData.expiresAt) {
                this.refreshTokens.delete(refreshPayload.jti);
                throw new Error('Refresh token expired');
            }

            // Mark refresh token as used if rotation is enabled
            if (this.config.enableRefreshTokenRotation) {
                refreshData.used = true;
                refreshData.usedAt = Date.now();
            }

            // Issue new access token
            const newAccessToken = await this.issueAccessToken({
                sub: refreshPayload.sub,
                // Include any additional claims from the original token
                ...options.additionalClaims
            }, options);

            // Issue new refresh token if rotation is enabled
            let newRefreshToken = null;
            if (this.config.enableRefreshTokenRotation) {
                newRefreshToken = await this.issueRefreshToken({
                    sub: refreshPayload.sub
                }, options);
                
                // Remove old refresh token
                this.refreshTokens.delete(refreshPayload.jti);
            }

            // Update metrics
            this.metrics.refreshTokensUsed++;

            // Audit log
            this.log('Access token refreshed', {
                userId: refreshPayload.sub,
                oldRefreshTokenId: refreshPayload.jti,
                newAccessTokenId: newAccessToken.tokenId,
                newRefreshTokenId: newRefreshToken?.tokenId
            });

            return {
                accessToken: newAccessToken,
                refreshToken: newRefreshToken
            };

        } catch (error) {
            this.metrics.validationErrors++;
            this.log('Token refresh failed', { 
                error: error.message 
            }, 'error');
            throw error;
        }
    }

    /**
     * Revoke token
     */
    async revokeToken(token, reason = 'user_request') {
        try {
            // Validate token to get payload
            const payload = await this.validateToken(token);

            // Add to blacklist
            if (this.config.enableTokenBlacklist) {
                this.tokenBlacklist.add(payload.jti);
            }

            // Remove from active tokens
            this.removeActiveToken(payload.sub, payload.jti);

            // If it's a refresh token, remove from storage
            if (payload.type === 'refresh') {
                this.refreshTokens.delete(payload.jti);
            }

            // Update metrics
            this.metrics.tokensRevoked++;

            // Audit log
            this.log('Token revoked', {
                userId: payload.sub,
                tokenId: payload.jti,
                type: payload.type,
                reason
            });

            return { success: true, tokenId: payload.jti };

        } catch (error) {
            this.log('Token revocation failed', { 
                error: error.message,
                reason 
            }, 'error');
            throw error;
        }
    }

    /**
     * Revoke all tokens for user
     */
    async revokeAllUserTokens(userId, reason = 'security_incident') {
        try {
            const activeTokens = this.activeTokens.get(userId) || new Set();
            const revokedTokens = [];

            // Add all active tokens to blacklist
            for (const tokenId of activeTokens) {
                if (this.config.enableTokenBlacklist) {
                    this.tokenBlacklist.add(tokenId);
                }
                revokedTokens.push(tokenId);
            }

            // Remove all refresh tokens for user
            for (const [tokenId, refreshData] of this.refreshTokens) {
                if (refreshData.userId === userId) {
                    this.refreshTokens.delete(tokenId);
                    revokedTokens.push(tokenId);
                }
            }

            // Clear active tokens
            this.activeTokens.delete(userId);

            // Update metrics
            this.metrics.tokensRevoked += revokedTokens.length;

            // Audit log
            this.log('All user tokens revoked', {
                userId,
                revokedCount: revokedTokens.length,
                reason
            });

            return { 
                success: true, 
                revokedCount: revokedTokens.length,
                revokedTokens 
            };

        } catch (error) {
            this.log('User token revocation failed', { 
                userId,
                error: error.message,
                reason 
            }, 'error');
            throw error;
        }
    }

    /**
     * Get token information
     */
    async getTokenInfo(token) {
        try {
            const payload = await this.validateToken(token);
            
            return {
                tokenId: payload.jti,
                userId: payload.sub,
                type: payload.type,
                issuer: payload.iss,
                audience: payload.aud,
                issuedAt: new Date(payload.iat * 1000),
                expiresAt: new Date(payload.exp * 1000),
                region: payload.region,
                tenant: payload.tenant,
                isExpired: Date.now() > payload.exp * 1000,
                isRevoked: this.tokenBlacklist.has(payload.jti)
            };

        } catch (error) {
            throw new Error('Invalid token');
        }
    }

    /**
     * Encrypt token (JWE)
     */
    async encryptToken(token) {
        if (!this.config.enableEncryption) {
            return token;
        }

        // Simplified JWE implementation
        // In production, use a proper JWE library
        const key = Array.from(this.encryptionKeys.values())[0]?.key;
        if (!key) {
            throw new Error('No encryption key available');
        }

        const cipher = crypto.createCipher('aes-256-gcm', key);
        let encrypted = cipher.update(token, 'utf8', 'hex');
        encrypted += cipher.final('hex');
        
        return `encrypted.${encrypted}`;
    }

    /**
     * Decrypt token (JWE)
     */
    async decryptToken(encryptedToken) {
        if (!this.config.enableEncryption || !encryptedToken.startsWith('encrypted.')) {
            return encryptedToken;
        }

        const key = Array.from(this.encryptionKeys.values())[0]?.key;
        if (!key) {
            throw new Error('No decryption key available');
        }

        const encrypted = encryptedToken.replace('encrypted.', '');
        const decipher = crypto.createDecipher('aes-256-gcm', key);
        let decrypted = decipher.update(encrypted, 'hex', 'utf8');
        decrypted += decipher.final('utf8');
        
        return decrypted;
    }

    /**
     * Utility methods
     */
    generateKeyId() {
        return crypto.randomBytes(8).toString('hex');
    }

    generateTokenId() {
        return crypto.randomBytes(16).toString('hex');
    }

    calculateExpiry(expiry) {
        if (typeof expiry === 'number') {
            return Math.floor(Date.now() / 1000) + expiry;
        }
        
        // Parse string expiry (e.g., '15m', '7d')
        const match = expiry.match(/^(\d+)([smhd])$/);
        if (!match) {
            throw new Error('Invalid expiry format');
        }

        const value = parseInt(match[1]);
        const unit = match[2];
        
        const multipliers = {
            s: 1,
            m: 60,
            h: 3600,
            d: 86400
        };

        return Math.floor(Date.now() / 1000) + (value * multipliers[unit]);
    }

    validateTokenPayload(payload) {
        if (!payload || typeof payload !== 'object') {
            throw new Error('Invalid token payload');
        }

        if (!payload.sub) {
            throw new Error('Token payload must include subject (sub)');
        }

        // Additional validation based on compliance requirements
        if (this.config.enableGDPRCompliance && payload.region === 'eu-central-1') {
            // GDPR-specific validation
            if (!payload.gdprConsent) {
                throw new Error('GDPR consent required for EU users');
            }
        }
    }

    trackActiveToken(userId, tokenId) {
        if (!this.activeTokens.has(userId)) {
            this.activeTokens.set(userId, new Set());
        }

        const userTokens = this.activeTokens.get(userId);
        userTokens.add(tokenId);

        // Enforce max tokens per user
        if (userTokens.size > this.config.maxTokensPerUser) {
            // Remove oldest token
            const oldestToken = userTokens.values().next().value;
            userTokens.delete(oldestToken);
            
            if (this.config.enableTokenBlacklist) {
                this.tokenBlacklist.add(oldestToken);
            }

            this.log('Token limit exceeded, oldest token revoked', {
                userId,
                revokedTokenId: oldestToken
            });
        }
    }

    removeActiveToken(userId, tokenId) {
        const userTokens = this.activeTokens.get(userId);
        if (userTokens) {
            userTokens.delete(tokenId);
            if (userTokens.size === 0) {
                this.activeTokens.delete(userId);
            }
        }
    }

    checkRateLimit(userId, type) {
        const key = `${userId}:${type}`;
        const now = Date.now();
        const hourAgo = now - 3600000; // 1 hour

        if (!this.rateLimitTracking.has(key)) {
            this.rateLimitTracking.set(key, []);
        }

        const requests = this.rateLimitTracking.get(key);
        
        // Remove old requests
        const recentRequests = requests.filter(timestamp => timestamp > hourAgo);
        this.rateLimitTracking.set(key, recentRequests);

        const limit = type === 'access' ? 
            this.config.maxTokenRequests : 
            this.config.maxRefreshRequests;

        return recentRequests.length < limit;
    }

    updateRateLimit(userId, type) {
        const key = `${userId}:${type}`;
        const requests = this.rateLimitTracking.get(key) || [];
        requests.push(Date.now());
        this.rateLimitTracking.set(key, requests);
    }

    shouldRotateKey() {
        if (!this.currentKeyId) return true;
        
        const currentKey = this.keyPairs.get(this.currentKeyId);
        if (!currentKey) return true;
        
        return Date.now() - currentKey.createdAt > this.config.keyRotationInterval;
    }

    setupKeyRotation() {
        setInterval(async () => {
            if (this.shouldRotateKey()) {
                try {
                    await this.generateKeyPair();
                    this.log('Key rotation completed', { 
                        newKeyId: this.currentKeyId 
                    });
                } catch (error) {
                    this.log('Key rotation failed', { 
                        error: error.message 
                    }, 'error');
                }
            }
        }, this.config.keyRotationInterval / 4); // Check every quarter of rotation interval
    }

    setupCleanupIntervals() {
        // Clean up expired tokens from blacklist every hour
        setInterval(() => {
            const now = Math.floor(Date.now() / 1000);
            const expiredTokens = [];

            for (const tokenId of this.tokenBlacklist) {
                // In a real implementation, you'd need to store expiry times
                // For now, we'll keep tokens in blacklist for a reasonable time
            }

            this.log('Blacklist cleanup completed', { 
                removedTokens: expiredTokens.length 
            });
        }, 3600000); // 1 hour

        // Clean up expired refresh tokens every hour
        setInterval(() => {
            const now = Date.now();
            const expiredTokens = [];

            for (const [tokenId, refreshData] of this.refreshTokens) {
                if (now > refreshData.expiresAt) {
                    this.refreshTokens.delete(tokenId);
                    expiredTokens.push(tokenId);
                }
            }

            this.log('Refresh token cleanup completed', { 
                removedTokens: expiredTokens.length 
            });
        }, 3600000); // 1 hour

        // Clean up old key pairs (keep last 3)
        setInterval(() => {
            const keyIds = Array.from(this.keyPairs.keys());
            if (keyIds.length > 3) {
                const sortedKeys = keyIds
                    .map(id => ({ id, createdAt: this.keyPairs.get(id).createdAt }))
                    .sort((a, b) => b.createdAt - a.createdAt);

                // Remove oldest keys, keep newest 3
                for (let i = 3; i < sortedKeys.length; i++) {
                    this.keyPairs.delete(sortedKeys[i].id);
                }

                this.log('Old key pairs cleaned up', { 
                    removedKeys: sortedKeys.length - 3 
                });
            }
        }, 86400000); // 24 hours
    }

    log(message, data = {}, level = 'info') {
        if (!this.config.enableAuditLogging) return;

        const logEntry = {
            timestamp: new Date().toISOString(),
            level,
            message,
            data,
            region: this.config.region
        };

        this.auditLogs.push(logEntry);

        // Keep only last 1000 log entries in memory
        if (this.auditLogs.length > 1000) {
            this.auditLogs.shift();
        }

        // In production, send to external logging service
        console.log(`[JWT-${level.toUpperCase()}] ${message}`, data);
    }

    /**
     * Get JWT manager statistics
     */
    getStats() {
        return {
            ...this.metrics,
            activeUsers: this.activeTokens.size,
            totalActiveTokens: Array.from(this.activeTokens.values())
                .reduce((sum, tokens) => sum + tokens.size, 0),
            blacklistedTokens: this.tokenBlacklist.size,
            refreshTokens: this.refreshTokens.size,
            keyPairs: this.keyPairs.size,
            currentKeyId: this.currentKeyId,
            encryptionEnabled: this.config.enableEncryption,
            region: this.config.region
        };
    }

    /**
     * Get public keys for token verification (JWKS endpoint)
     */
    getPublicKeys() {
        const keys = [];

        for (const [keyId, keyPair] of this.keyPairs) {
            // Convert PEM to JWK format (simplified)
            keys.push({
                kty: 'RSA',
                kid: keyId,
                use: 'sig',
                alg: keyPair.algorithm,
                // In production, properly convert PEM to JWK
                n: 'base64url-encoded-modulus',
                e: 'AQAB'
            });
        }

        return { keys };
    }

    /**
     * Health check
     */
    healthCheck() {
        const issues = [];

        // Check if we have valid keys
        if (this.keyPairs.size === 0) {
            issues.push('No signing keys available');
        }

        // Check if current key is valid
        if (!this.currentKeyId || !this.keyPairs.has(this.currentKeyId)) {
            issues.push('No current signing key');
        }

        // Check if encryption is properly configured
        if (this.config.enableEncryption && this.encryptionKeys.size === 0) {
            issues.push('Encryption enabled but no encryption keys available');
        }

        return {
            healthy: issues.length === 0,
            issues,
            stats: this.getStats()
        };
    }
}

export default JWTManager;