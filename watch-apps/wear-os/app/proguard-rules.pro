# Add project specific ProGuard rules here.
# You can control the set of applied configuration files using the
# proguardFiles setting in build.gradle.
#
# For more details, see
#   http://developer.android.com/guide/developing/tools/proguard.html

# If your project uses WebView with JS, uncomment the following
# and specify the fully qualified class name to the JavaScript interface
# class:
#-keepclassmembers class fqcn.of.javascript.interface.for.webview {
#   public *;
#}

# Uncomment this to preserve the line number information for
# debugging stack traces.
-keepattributes SourceFile,LineNumberTable

# If you keep the line number information, uncomment this to
# hide the original source file name.
-renamesourcefileattribute SourceFile

# Keep all classes in our package
-keep class com.taishanglaojun.wearos.** { *; }

# Keep all service classes
-keep class * extends android.app.Service { *; }

# Keep all data model classes
-keep class com.taishanglaojun.wearos.models.** { *; }

# Keep all utility classes
-keep class com.taishanglaojun.wearos.utils.** { *; }

# Keep performance optimizer
-keep class com.taishanglaojun.wearos.utils.PerformanceOptimizer { *; }
-keep class com.taishanglaojun.wearos.utils.PerformanceOptimizer$* { *; }

# Keep WearOS data model
-keep class com.taishanglaojun.wearos.models.WearOSData { *; }
-keep class com.taishanglaojun.wearos.models.WearOSData$* { *; }

# Keep service classes
-keep class com.taishanglaojun.wearos.services.LocationService { *; }
-keep class com.taishanglaojun.wearos.services.HealthService { *; }
-keep class com.taishanglaojun.wearos.services.DataSyncService { *; }

# Keep MainActivity
-keep class com.taishanglaojun.wearos.MainActivity { *; }

# Keep all classes that have native methods
-keepclasseswithmembernames class * {
    native <methods>;
}

# Keep all classes that are used by reflection
-keepclassmembers class * {
    @android.webkit.JavascriptInterface <methods>;
}

# Keep all enums
-keepclassmembers enum * {
    public static **[] values();
    public static ** valueOf(java.lang.String);
}

# Keep all Parcelable implementations
-keep class * implements android.os.Parcelable {
    public static final android.os.Parcelable$Creator *;
}

# Keep all Serializable classes
-keepclassmembers class * implements java.io.Serializable {
    static final long serialVersionUID;
    private static final java.io.ObjectStreamField[] serialPersistentFields;
    private void writeObject(java.io.ObjectOutputStream);
    private void readObject(java.io.ObjectInputStream);
    java.lang.Object writeReplace();
    java.lang.Object readResolve();
}

# Keep all classes with @Keep annotation
-keep @androidx.annotation.Keep class * { *; }
-keepclassmembers class * {
    @androidx.annotation.Keep *;
}

# Keep all classes with @DoNotStrip annotation
-keep @com.facebook.proguard.annotations.DoNotStrip class * { *; }
-keepclassmembers class * {
    @com.facebook.proguard.annotations.DoNotStrip *;
}

# Gson specific rules
-keepattributes Signature
-keepattributes *Annotation*
-dontwarn sun.misc.**
-keep class com.google.gson.** { *; }
-keep class * implements com.google.gson.TypeAdapterFactory
-keep class * implements com.google.gson.JsonSerializer
-keep class * implements com.google.gson.JsonDeserializer

# Retrofit specific rules
-keepattributes Signature, InnerClasses, EnclosingMethod
-keepattributes RuntimeVisibleAnnotations, RuntimeVisibleParameterAnnotations
-keepclassmembers,allowshrinking,allowobfuscation interface * {
    @retrofit2.http.* <methods>;
}
-dontwarn org.codehaus.mojo.animal_sniffer.IgnoreJRERequirement
-dontwarn javax.annotation.**
-dontwarn kotlin.Unit
-dontwarn retrofit2.KotlinExtensions
-dontwarn retrofit2.KotlinExtensions$*

# OkHttp specific rules
-dontwarn okhttp3.**
-dontwarn okio.**
-dontwarn javax.annotation.**
-keepnames class okhttp3.internal.publicsuffix.PublicSuffixDatabase

# Wear OS specific rules
-keep class androidx.wear.** { *; }
-keep class com.google.android.wearable.** { *; }
-keep class com.google.android.support.wearable.** { *; }

# Health Services specific rules
-keep class androidx.health.** { *; }
-dontwarn androidx.health.**

# Location Services specific rules
-keep class com.google.android.gms.location.** { *; }
-dontwarn com.google.android.gms.**

# Remove logging in release builds
-assumenosideeffects class android.util.Log {
    public static boolean isLoggable(java.lang.String, int);
    public static int v(...);
    public static int i(...);
    public static int w(...);
    public static int d(...);
    public static int e(...);
}

# Optimize and obfuscate
-optimizations !code/simplification/arithmetic,!code/simplification/cast,!field/*,!class/merging/*
-optimizationpasses 5
-allowaccessmodification
-dontpreverify
-repackageclasses ''
-allowaccessmodification
-keepattributes *Annotation*