import { cryptoFramework } from '@kit.CryptoArchitectureKit';
import { preferences } from '@kit.ArkData';
import { LocationPoint } from '../data/LocationPoint';
import { Trajectory } from '../data/Trajectory';

/**
 * 加密数据结构
 */
export interface EncryptedData {
  encryptedData: string;
  iv: string;
}

/**
 * 加密的位置点
 */
export interface EncryptedLocationPoint {
  id: string;
  trajectoryId: string;
  timestamp: number;
  encryptedData: string;
  iv: string;
  checksum: string;
}

/**
 * 加密的轨迹
 */
export interface EncryptedTrajectory {
  id: string;
  name: string;
  startTime?: number;
  endTime?: number;
  encryptedData: string;
  iv: string;
  checksum: string;
}

/**
 * 传输加密数据
 */
export interface TransportEncryptedData {
  encryptedData: string;
  iv: string;
  timestamp: number;
  checksum: string;
}

/**
 * 数据加密管理器
 * 负责位置数据的加密存储和安全传输
 */
export class DataEncryption {
  private static instance: DataEncryption;
  private static readonly PREF_NAME = 'encrypted_prefs';
  private static readonly KEY_SIZE = 256;
  private static readonly IV_SIZE = 16;
  private static readonly TAG_SIZE = 16;
  
  private encryptedPrefs: preferences.Preferences | null = null;
  private symmetricKey: cryptoFramework.SymKey | null = null;
  
  private constructor() {
    this.initializeEncryption();
  }
  
  /**
   * 获取单例实例
   */
  public static getInstance(): DataEncryption {
    if (!DataEncryption.instance) {
      DataEncryption.instance = new DataEncryption();
    }
    return DataEncryption.instance;
  }
  
  /**
   * 初始化加密组件
   */
  private async initializeEncryption(): Promise<void> {
    try {
      // 初始化加密首选项
      this.encryptedPrefs = await preferences.getPreferences(getContext(), DataEncryption.PREF_NAME);
      
      // 生成或获取对称密钥
      await this.generateOrGetSymmetricKey();
    } catch (error) {
      console.error('Failed to initialize encryption:', error);
      throw new Error('Encryption initialization failed');
    }
  }
  
  /**
   * 生成或获取对称密钥
   */
  private async generateOrGetSymmetricKey(): Promise<void> {
    try {
      const storedKey = await this.getSecureConfig('symmetric_key');
      
      if (storedKey) {
        // 从存储的密钥数据恢复密钥
        const keyData = new Uint8Array(JSON.parse(storedKey));
        const symKeyGenerator = cryptoFramework.createSymKeyGenerator('AES256');
        this.symmetricKey = await symKeyGenerator.convertKey(keyData);
      } else {
        // 生成新的对称密钥
        const symKeyGenerator = cryptoFramework.createSymKeyGenerator('AES256');
        this.symmetricKey = await symKeyGenerator.generateSymKey();
        
        // 存储密钥数据
        const keyData = this.symmetricKey.getEncoded();
        await this.storeSecureConfig('symmetric_key', JSON.stringify(Array.from(keyData)));
      }
    } catch (error) {
      console.error('Failed to generate or get symmetric key:', error);
      throw new Error('Key generation failed');
    }
  }
  
  /**
   * 加密数据
   */
  public async encryptData(data: string): Promise<EncryptedData> {
    try {
      if (!this.symmetricKey) {
        throw new Error('Symmetric key not initialized');
      }
      
      const cipher = cryptoFramework.createCipher('AES256|GCM|PKCS7');
      await cipher.init(cryptoFramework.CryptoMode.ENCRYPT_MODE, this.symmetricKey, null);
      
      const plainTextData = new Uint8Array(Buffer.from(data, 'utf-8'));
      const encryptedData = await cipher.doFinal(plainTextData);
      
      // 获取IV和加密数据
      const iv = cipher.getSpec().iv;
      
      return {
        encryptedData: Buffer.from(encryptedData).toString('base64'),
        iv: Buffer.from(iv).toString('base64')
      };
    } catch (error) {
      console.error('Failed to encrypt data:', error);
      throw new Error('Data encryption failed');
    }
  }
  
  /**
   * 解密数据
   */
  public async decryptData(encryptedData: EncryptedData): Promise<string> {
    try {
      if (!this.symmetricKey) {
        throw new Error('Symmetric key not initialized');
      }
      
      const cipher = cryptoFramework.createCipher('AES256|GCM|PKCS7');
      const iv = new Uint8Array(Buffer.from(encryptedData.iv, 'base64'));
      const gcmParamsSpec: cryptoFramework.GcmParamsSpec = {
        iv: iv,
        aad: new Uint8Array(0),
        authTag: new Uint8Array(DataEncryption.TAG_SIZE)
      };
      
      await cipher.init(cryptoFramework.CryptoMode.DECRYPT_MODE, this.symmetricKey, gcmParamsSpec);
      
      const cipherData = new Uint8Array(Buffer.from(encryptedData.encryptedData, 'base64'));
      const decryptedData = await cipher.doFinal(cipherData);
      
      return Buffer.from(decryptedData).toString('utf-8');
    } catch (error) {
      console.error('Failed to decrypt data:', error);
      throw new Error('Data decryption failed');
    }
  }
  
  /**
   * 加密位置点
   */
  public async encryptLocationPoint(locationPoint: LocationPoint): Promise<EncryptedLocationPoint> {
    try {
      const json = JSON.stringify(locationPoint);
      const encryptedData = await this.encryptData(json);
      
      return {
        id: locationPoint.id,
        trajectoryId: locationPoint.trajectoryId,
        timestamp: locationPoint.timestamp,
        encryptedData: encryptedData.encryptedData,
        iv: encryptedData.iv,
        checksum: await this.calculateChecksum(json)
      };
    } catch (error) {
      console.error('Failed to encrypt location point:', error);
      throw new Error('Location point encryption failed');
    }
  }
  
  /**
   * 解密位置点
   */
  public async decryptLocationPoint(encryptedLocationPoint: EncryptedLocationPoint): Promise<LocationPoint> {
    try {
      const encryptedData: EncryptedData = {
        encryptedData: encryptedLocationPoint.encryptedData,
        iv: encryptedLocationPoint.iv
      };
      
      const json = await this.decryptData(encryptedData);
      
      // 验证校验和
      const calculatedChecksum = await this.calculateChecksum(json);
      if (calculatedChecksum !== encryptedLocationPoint.checksum) {
        throw new Error('Data integrity check failed');
      }
      
      return JSON.parse(json) as LocationPoint;
    } catch (error) {
      console.error('Failed to decrypt location point:', error);
      throw new Error('Location point decryption failed');
    }
  }
  
  /**
   * 加密轨迹
   */
  public async encryptTrajectory(trajectory: Trajectory): Promise<EncryptedTrajectory> {
    try {
      const json = JSON.stringify(trajectory);
      const encryptedData = await this.encryptData(json);
      
      return {
        id: trajectory.id,
        name: trajectory.name,
        startTime: trajectory.startTime,
        endTime: trajectory.endTime,
        encryptedData: encryptedData.encryptedData,
        iv: encryptedData.iv,
        checksum: await this.calculateChecksum(json)
      };
    } catch (error) {
      console.error('Failed to encrypt trajectory:', error);
      throw new Error('Trajectory encryption failed');
    }
  }
  
  /**
   * 解密轨迹
   */
  public async decryptTrajectory(encryptedTrajectory: EncryptedTrajectory): Promise<Trajectory> {
    try {
      const encryptedData: EncryptedData = {
        encryptedData: encryptedTrajectory.encryptedData,
        iv: encryptedTrajectory.iv
      };
      
      const json = await this.decryptData(encryptedData);
      
      // 验证校验和
      const calculatedChecksum = await this.calculateChecksum(json);
      if (calculatedChecksum !== encryptedTrajectory.checksum) {
        throw new Error('Data integrity check failed');
      }
      
      return JSON.parse(json) as Trajectory;
    } catch (error) {
      console.error('Failed to decrypt trajectory:', error);
      throw new Error('Trajectory decryption failed');
    }
  }
  
  /**
   * 安全存储敏感配置
   */
  public async storeSecureConfig(key: string, value: string): Promise<void> {
    try {
      if (!this.encryptedPrefs) {
        throw new Error('Encrypted preferences not initialized');
      }
      
      await this.encryptedPrefs.put(key, value);
      await this.encryptedPrefs.flush();
    } catch (error) {
      console.error('Failed to store secure config:', error);
      throw new Error('Secure config storage failed');
    }
  }
  
  /**
   * 获取安全存储的配置
   */
  public async getSecureConfig(key: string, defaultValue?: string): Promise<string | null> {
    try {
      if (!this.encryptedPrefs) {
        throw new Error('Encrypted preferences not initialized');
      }
      
      const value = await this.encryptedPrefs.get(key, defaultValue || '');
      return value as string || null;
    } catch (error) {
      console.error('Failed to get secure config:', error);
      return null;
    }
  }
  
  /**
   * 删除安全存储的配置
   */
  public async removeSecureConfig(key: string): Promise<void> {
    try {
      if (!this.encryptedPrefs) {
        throw new Error('Encrypted preferences not initialized');
      }
      
      await this.encryptedPrefs.delete(key);
      await this.encryptedPrefs.flush();
    } catch (error) {
      console.error('Failed to remove secure config:', error);
      throw new Error('Secure config removal failed');
    }
  }
  
  /**
   * 生成传输用的临时密钥
   */
  public generateTransportKey(): string {
    const keyBytes = new Uint8Array(32); // 256位密钥
    crypto.getRandomValues(keyBytes);
    return Buffer.from(keyBytes).toString('base64');
  }
  
  /**
   * 使用传输密钥加密数据
   */
  public async encryptForTransport(data: string, transportKey: string): Promise<TransportEncryptedData> {
    try {
      const keyBytes = new Uint8Array(Buffer.from(transportKey, 'base64'));
      const symKeyGenerator = cryptoFramework.createSymKeyGenerator('AES256');
      const secretKey = await symKeyGenerator.convertKey(keyBytes);
      
      const cipher = cryptoFramework.createCipher('AES256|GCM|PKCS7');
      await cipher.init(cryptoFramework.CryptoMode.ENCRYPT_MODE, secretKey, null);
      
      const plainTextData = new Uint8Array(Buffer.from(data, 'utf-8'));
      const encryptedData = await cipher.doFinal(plainTextData);
      const iv = cipher.getSpec().iv;
      
      return {
        encryptedData: Buffer.from(encryptedData).toString('base64'),
        iv: Buffer.from(iv).toString('base64'),
        timestamp: Date.now(),
        checksum: await this.calculateChecksum(data)
      };
    } catch (error) {
      console.error('Failed to encrypt for transport:', error);
      throw new Error('Transport encryption failed');
    }
  }
  
  /**
   * 使用传输密钥解密数据
   */
  public async decryptFromTransport(transportData: TransportEncryptedData, transportKey: string): Promise<string> {
    try {
      const keyBytes = new Uint8Array(Buffer.from(transportKey, 'base64'));
      const symKeyGenerator = cryptoFramework.createSymKeyGenerator('AES256');
      const secretKey = await symKeyGenerator.convertKey(keyBytes);
      
      const cipher = cryptoFramework.createCipher('AES256|GCM|PKCS7');
      const iv = new Uint8Array(Buffer.from(transportData.iv, 'base64'));
      const gcmParamsSpec: cryptoFramework.GcmParamsSpec = {
        iv: iv,
        aad: new Uint8Array(0),
        authTag: new Uint8Array(DataEncryption.TAG_SIZE)
      };
      
      await cipher.init(cryptoFramework.CryptoMode.DECRYPT_MODE, secretKey, gcmParamsSpec);
      
      const cipherData = new Uint8Array(Buffer.from(transportData.encryptedData, 'base64'));
      const decryptedData = await cipher.doFinal(cipherData);
      const decryptedString = Buffer.from(decryptedData).toString('utf-8');
      
      // 验证校验和
      const calculatedChecksum = await this.calculateChecksum(decryptedString);
      if (calculatedChecksum !== transportData.checksum) {
        throw new Error('Transport data integrity check failed');
      }
      
      return decryptedString;
    } catch (error) {
      console.error('Failed to decrypt from transport:', error);
      throw new Error('Transport decryption failed');
    }
  }
  
  /**
   * 计算数据校验和
   */
  private async calculateChecksum(data: string): Promise<string> {
    try {
      const md = cryptoFramework.createMd('SHA256');
      await md.update(new Uint8Array(Buffer.from(data, 'utf-8')));
      const hashBytes = await md.digest();
      return Buffer.from(hashBytes).toString('base64').substring(0, 16);
    } catch (error) {
      console.error('Failed to calculate checksum:', error);
      throw new Error('Checksum calculation failed');
    }
  }
  
  /**
   * 清除所有加密数据
   */
  public async clearAllEncryptedData(): Promise<void> {
    try {
      if (this.encryptedPrefs) {
        await this.encryptedPrefs.clear();
        await this.encryptedPrefs.flush();
      }
      
      // 重新生成密钥
      await this.generateOrGetSymmetricKey();
    } catch (error) {
      console.error('Failed to clear encrypted data:', error);
      throw new Error('Clear encrypted data failed');
    }
  }
  
  /**
   * 验证数据完整性
   */
  public async verifyDataIntegrity(originalData: string, checksum: string): Promise<boolean> {
    try {
      const calculatedChecksum = await this.calculateChecksum(originalData);
      return calculatedChecksum === checksum;
    } catch (error) {
      console.error('Failed to verify data integrity:', error);
      return false;
    }
  }
}