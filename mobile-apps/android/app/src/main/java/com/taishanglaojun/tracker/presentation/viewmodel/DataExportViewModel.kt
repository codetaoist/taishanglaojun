package com.taishanglaojun.tracker.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.taishanglaojun.tracker.data.service.DataExportService
import com.taishanglaojun.tracker.data.service.ExportResult
import com.taishanglaojun.tracker.data.service.ImportResult
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import java.io.File
import javax.inject.Inject

@HiltViewModel
class DataExportViewModel @Inject constructor(
    private val dataExportService: DataExportService
) : ViewModel() {
    
    private val _uiState = MutableStateFlow(DataExportUiState())
    val uiState: StateFlow<DataExportUiState> = _uiState.asStateFlow()
    
    fun exportAllData(outputDir: File) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(
                isExporting = true,
                exportProgress = 0f,
                error = null
            )
            
            try {
                // 模拟进度更新
                _uiState.value = _uiState.value.copy(exportProgress = 0.2f)
                
                val result = dataExportService.exportAllData(outputDir)
                
                _uiState.value = _uiState.value.copy(exportProgress = 0.8f)
                
                if (result.isSuccess) {
                    val exportResult = result.getOrThrow()
                    _uiState.value = _uiState.value.copy(
                        isExporting = false,
                        exportProgress = 1.0f,
                        lastExportResult = exportResult,
                        exportSuccess = true
                    )
                } else {
                    _uiState.value = _uiState.value.copy(
                        isExporting = false,
                        exportProgress = 0f,
                        error = result.exceptionOrNull()?.message ?: "导出失败",
                        exportSuccess = false
                    )
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isExporting = false,
                    exportProgress = 0f,
                    error = e.message ?: "导出过程中发生错误",
                    exportSuccess = false
                )
            }
        }
    }
    
    fun exportToZip(outputDir: File) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(
                isExporting = true,
                exportProgress = 0f,
                error = null
            )
            
            try {
                _uiState.value = _uiState.value.copy(exportProgress = 0.3f)
                
                val result = dataExportService.exportToZip(outputDir)
                
                _uiState.value = _uiState.value.copy(exportProgress = 0.9f)
                
                if (result.isSuccess) {
                    val zipFile = result.getOrThrow()
                    _uiState.value = _uiState.value.copy(
                        isExporting = false,
                        exportProgress = 1.0f,
                        lastExportFile = zipFile,
                        exportSuccess = true
                    )
                } else {
                    _uiState.value = _uiState.value.copy(
                        isExporting = false,
                        exportProgress = 0f,
                        error = result.exceptionOrNull()?.message ?: "ZIP导出失败",
                        exportSuccess = false
                    )
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isExporting = false,
                    exportProgress = 0f,
                    error = e.message ?: "ZIP导出过程中发生错误",
                    exportSuccess = false
                )
            }
        }
    }
    
    fun importData(importFile: File) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(
                isImporting = true,
                importProgress = 0f,
                error = null
            )
            
            try {
                _uiState.value = _uiState.value.copy(importProgress = 0.3f)
                
                val result = dataExportService.importData(importFile)
                
                _uiState.value = _uiState.value.copy(importProgress = 0.8f)
                
                if (result.isSuccess) {
                    val importResult = result.getOrThrow()
                    _uiState.value = _uiState.value.copy(
                        isImporting = false,
                        importProgress = 1.0f,
                        lastImportResult = importResult,
                        importSuccess = true
                    )
                } else {
                    _uiState.value = _uiState.value.copy(
                        isImporting = false,
                        importProgress = 0f,
                        error = result.exceptionOrNull()?.message ?: "导入失败",
                        importSuccess = false
                    )
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isImporting = false,
                    importProgress = 0f,
                    error = e.message ?: "导入过程中发生错误",
                    importSuccess = false
                )
            }
        }
    }
    
    fun clearError() {
        _uiState.value = _uiState.value.copy(error = null)
    }
    
    fun resetExportState() {
        _uiState.value = _uiState.value.copy(
            exportSuccess = false,
            exportProgress = 0f,
            lastExportResult = null,
            lastExportFile = null
        )
    }
    
    fun resetImportState() {
        _uiState.value = _uiState.value.copy(
            importSuccess = false,
            importProgress = 0f,
            lastImportResult = null
        )
    }
}

data class DataExportUiState(
    val isExporting: Boolean = false,
    val isImporting: Boolean = false,
    val exportProgress: Float = 0f,
    val importProgress: Float = 0f,
    val exportSuccess: Boolean = false,
    val importSuccess: Boolean = false,
    val error: String? = null,
    val lastExportResult: ExportResult? = null,
    val lastImportResult: ImportResult? = null,
    val lastExportFile: File? = null
)