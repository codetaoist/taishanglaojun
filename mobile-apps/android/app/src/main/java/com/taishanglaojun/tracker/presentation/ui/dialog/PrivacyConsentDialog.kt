package com.taishanglaojun.tracker.presentation.ui.dialog

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import com.taishanglaojun.tracker.privacy.PrivacyManager

/**
 * 隐私同意对话框
 */
@Composable
fun PrivacyConsentDialog(
    onAccept: (PrivacyManager.PrivacyConsent) -> Unit,
    onDecline: () -> Unit
) {
    val context = LocalContext.current
    val privacyManager = remember { PrivacyManager.getInstance(context) }
    
    var dataCollectionConsent by remember { mutableStateOf(false) }
    var dataSharingConsent by remember { mutableStateOf(false) }
    var analyticsConsent by remember { mutableStateOf(false) }
    var showPrivacyPolicy by remember { mutableStateOf(false) }
    
    Dialog(
        onDismissRequest = { /* 不允许点击外部关闭 */ },
        properties = DialogProperties(
            dismissOnBackPress = false,
            dismissOnClickOutside = false
        )
    ) {
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            elevation = CardDefaults.cardElevation(defaultElevation = 8.dp)
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(24.dp)
            ) {
                // 标题
                Text(
                    text = "隐私政策与用户协议",
                    fontSize = 20.sp,
                    fontWeight = FontWeight.Bold,
                    textAlign = TextAlign.Center,
                    modifier = Modifier.fillMaxWidth()
                )
                
                Spacer(modifier = Modifier.height(16.dp))
                
                // 说明文字
                Text(
                    text = "为了为您提供更好的服务，我们需要获得您的同意来收集和使用相关数据。请仔细阅读以下条款：",
                    fontSize = 14.sp,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                
                Spacer(modifier = Modifier.height(16.dp))
                
                // 同意选项
                ConsentCheckbox(
                    text = "我同意收集位置数据用于轨迹记录",
                    checked = dataCollectionConsent,
                    onCheckedChange = { dataCollectionConsent = it },
                    required = true
                )
                
                ConsentCheckbox(
                    text = "我同意与云端服务同步数据（可选）",
                    checked = dataSharingConsent,
                    onCheckedChange = { dataSharingConsent = it },
                    required = false
                )
                
                ConsentCheckbox(
                    text = "我同意收集匿名使用统计数据（可选）",
                    checked = analyticsConsent,
                    onCheckedChange = { analyticsConsent = it },
                    required = false
                )
                
                Spacer(modifier = Modifier.height(16.dp))
                
                // 隐私政策链接
                TextButton(
                    onClick = { showPrivacyPolicy = true },
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Text(
                        text = "查看完整隐私政策",
                        color = MaterialTheme.colorScheme.primary
                    )
                }
                
                Spacer(modifier = Modifier.height(24.dp))
                
                // 按钮
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    OutlinedButton(
                        onClick = onDecline,
                        modifier = Modifier.weight(1f)
                    ) {
                        Text("拒绝")
                    }
                    
                    Button(
                        onClick = {
                            val consent = PrivacyManager.PrivacyConsent(
                                privacyPolicyAccepted = true,
                                dataCollectionConsent = dataCollectionConsent,
                                dataSharingConsent = dataSharingConsent,
                                analyticsConsent = analyticsConsent
                            )
                            onAccept(consent)
                        },
                        enabled = dataCollectionConsent, // 必须同意数据收集
                        modifier = Modifier.weight(1f)
                    ) {
                        Text("同意并继续")
                    }
                }
            }
        }
    }
    
    // 隐私政策详情对话框
    if (showPrivacyPolicy) {
        PrivacyPolicyDialog(
            privacyPolicyText = privacyManager.getPrivacyPolicyText(),
            onDismiss = { showPrivacyPolicy = false }
        )
    }
}

/**
 * 同意复选框组件
 */
@Composable
private fun ConsentCheckbox(
    text: String,
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit,
    required: Boolean = false
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Checkbox(
            checked = checked,
            onCheckedChange = onCheckedChange
        )
        
        Spacer(modifier = Modifier.width(8.dp))
        
        Text(
            text = if (required) "$text *" else text,
            fontSize = 14.sp,
            color = if (required) MaterialTheme.colorScheme.onSurface 
                   else MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.weight(1f)
        )
    }
}

/**
 * 隐私政策详情对话框
 */
@Composable
private fun PrivacyPolicyDialog(
    privacyPolicyText: String,
    onDismiss: () -> Unit
) {
    Dialog(onDismissRequest = onDismiss) {
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .fillMaxHeight(0.8f)
                .padding(16.dp),
            elevation = CardDefaults.cardElevation(defaultElevation = 8.dp)
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(24.dp)
            ) {
                // 标题
                Text(
                    text = "隐私政策",
                    fontSize = 20.sp,
                    fontWeight = FontWeight.Bold,
                    textAlign = TextAlign.Center,
                    modifier = Modifier.fillMaxWidth()
                )
                
                Spacer(modifier = Modifier.height(16.dp))
                
                // 内容
                Text(
                    text = privacyPolicyText,
                    fontSize = 14.sp,
                    modifier = Modifier
                        .weight(1f)
                        .verticalScroll(rememberScrollState())
                )
                
                Spacer(modifier = Modifier.height(16.dp))
                
                // 关闭按钮
                Button(
                    onClick = onDismiss,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Text("关闭")
                }
            }
        }
    }
}

/**
 * 权限说明对话框
 */
@Composable
fun PermissionRationaleDialog(
    title: String,
    message: String,
    onConfirm: () -> Unit,
    onDismiss: () -> Unit
) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = {
            Text(
                text = title,
                fontWeight = FontWeight.Bold
            )
        },
        text = {
            Text(text = message)
        },
        confirmButton = {
            TextButton(onClick = onConfirm) {
                Text("确定")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("取消")
            }
        }
    )
}

/**
 * 数据脱敏设置对话框
 */
@Composable
fun DataSensitivityDialog(
    currentLevel: PrivacyManager.DataSensitivityLevel,
    onLevelSelected: (PrivacyManager.DataSensitivityLevel) -> Unit,
    onDismiss: () -> Unit
) {
    var selectedLevel by remember { mutableStateOf(currentLevel) }
    
    AlertDialog(
        onDismissRequest = onDismiss,
        title = {
            Text(
                text = "数据脱敏设置",
                fontWeight = FontWeight.Bold
            )
        },
        text = {
            Column {
                Text(
                    text = "选择位置数据的脱敏级别：",
                    modifier = Modifier.padding(bottom = 16.dp)
                )
                
                PrivacyManager.DataSensitivityLevel.values().forEach { level ->
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(vertical = 4.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        RadioButton(
                            selected = selectedLevel == level,
                            onClick = { selectedLevel = level }
                        )
                        
                        Spacer(modifier = Modifier.width(8.dp))
                        
                        Column {
                            Text(
                                text = when (level) {
                                    PrivacyManager.DataSensitivityLevel.NONE -> "无脱敏"
                                    PrivacyManager.DataSensitivityLevel.LOW -> "低级脱敏"
                                    PrivacyManager.DataSensitivityLevel.MEDIUM -> "中级脱敏"
                                    PrivacyManager.DataSensitivityLevel.HIGH -> "高级脱敏"
                                },
                                fontWeight = FontWeight.Medium
                            )
                            
                            Text(
                                text = when (level) {
                                    PrivacyManager.DataSensitivityLevel.NONE -> "保留完整精度"
                                    PrivacyManager.DataSensitivityLevel.LOW -> "保留4位小数（约11米精度）"
                                    PrivacyManager.DataSensitivityLevel.MEDIUM -> "保留3位小数（约111米精度）"
                                    PrivacyManager.DataSensitivityLevel.HIGH -> "保留2位小数（约1.1公里精度）"
                                },
                                fontSize = 12.sp,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    }
                }
            }
        },
        confirmButton = {
            TextButton(
                onClick = {
                    onLevelSelected(selectedLevel)
                    onDismiss()
                }
            ) {
                Text("确定")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("取消")
            }
        }
    )
}