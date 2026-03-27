import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { useFocusTrap } from '../../hooks/useFocusTrap';
import { X, CreditCard, QrCode, ArrowLeft, Copy, Check } from 'lucide-react';
import { copyToClipboard } from '../../utils/clipboard';

const DepositModal = ({ isOpen, onClose, onDepositSuccess }) => {
  const modalRef = useFocusTrap(isOpen);
  const [currentStep, setCurrentStep] = useState(1); // 1: Valor, 2: Método, 3: QR Code
  const [depositAmount, setDepositAmount] = useState('');
  const [selectedMethod, setSelectedMethod] = useState('pix');
  const [copied, setCopied] = useState(false);

  // Gerar QR Code fake para demonstração
  const fakePixCode = "BR.GOV.BCB.PIX 0102650017BR.GOV.BCB.PIX014012345678901234567890123456789012345678905204000053039865405100.005802BR5925PIXELCRAFT STUDIO LTDA6009SAO PAULO62070503***6304A1B2";

  const styles = {
    overlay: {
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      background: 'rgba(0, 0, 0, 0.8)',
      backdropFilter: 'blur(20px)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 1000,
      padding: 'var(--btn-padding-md)',
    },
    modal: {
      width: '100%',
      maxWidth: '500px',
      background: 'rgba(10, 14, 26, 0.95)',
      backdropFilter: 'blur(20px)',
      border: '1px solid rgba(88, 58, 255, 0.3)',
      borderRadius: '1.5rem',
      padding: '2rem',
      position: 'relative',
      boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5), inset 0 0 20px rgba(88, 58, 255, 0.2)',
    },
    closeButton: {
      position: 'absolute',
      top: '1.5rem',
      right: '1.5rem',
      background: 'none',
      border: 'none',
      color: '#B8BDC7',
      cursor: 'pointer',
      width: '36px',
      height: '36px',
      borderRadius: '50%',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      transition: 'all 0.3s ease',
    },
    header: {
      textAlign: 'center',
      marginBottom: '2rem',
    },
    title: {
      fontSize: 'var(--title-h4)',
      fontWeight: 700,
      background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 50%, #1AD2FF 100%)',
      WebkitBackgroundClip: 'text',
      WebkitTextFillColor: 'transparent',
      marginBottom: '0.5rem',
    },
    subtitle: {
      color: '#B8BDC7',
      fontSize: '1rem',
    },
    stepIndicator: {
      display: 'flex',
      justifyContent: 'center',
      gap: '0.5rem',
      marginBottom: '2rem',
    },
    stepDot: {
      width: '8px',
      height: '8px',
      borderRadius: '50%',
      background: '#B8BDC7',
    },
    stepDotActive: {
      background: 'var(--gradient-primary)',
    },
    inputField: {
      width: '100%',
      padding: '1rem 1.25rem',
      background: 'rgba(21, 26, 38, 0.8)',
      border: '1px solid rgba(88, 58, 255, 0.2)',
      borderRadius: '1rem',
      color: '#F8F9FA',
      fontSize: '1.1rem',
      fontWeight: 500,
      marginBottom: '1rem',
      outline: 'none',
      transition: 'all 0.3s ease',
    },
    inputFieldFocus: {
      border: '1px solid rgba(88, 58, 255, 0.4)',
      boxShadow: '0 0 20px rgba(88, 58, 255, 0.3)',
    },
    buttonPrimary: {
      width: '100%',
      background: 'var(--gradient-primary)',
      border: 'none',
      borderRadius: '1rem',
      padding: '1.25rem',
      color: '#F8F9FA',
      fontSize: '1.1rem',
      fontWeight: 700,
      cursor: 'pointer',
      transition: 'all 0.3s ease',
      boxShadow: '0 8px 32px rgba(88, 58, 255, 0.4)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '0.75rem',
      marginTop: '1rem',
    },
    buttonSecondary: {
      width: '100%',
      background: 'rgba(21, 26, 38, 0.6)',
      backdropFilter: 'blur(10px)',
      border: '1px solid rgba(88, 58, 255, 0.2)',
      borderRadius: '1rem',
      padding: '1.25rem',
      color: '#F8F9FA',
      fontSize: '1.1rem',
      fontWeight: 700,
      cursor: 'pointer',
      transition: 'all 0.3s ease',
      boxShadow: '0 8px 32px rgba(0, 0, 0, 0.3)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '0.75rem',
      marginTop: '1rem',
    },
    paymentMethod: {
      display: 'flex',
      alignItems: 'center',
      gap: '1rem',
      padding: 'var(--btn-padding-md)',
      background: 'rgba(21, 26, 38, 0.6)',
      border: '1px solid rgba(88, 58, 255, 0.2)',
      borderRadius: '1rem',
      cursor: 'pointer',
      transition: 'all 0.3s ease',
      marginBottom: '1rem',
    },
    paymentMethodSelected: {
      background: 'rgba(88, 58, 255, 0.15)',
      border: '1px solid rgba(88, 58, 255, 0.4)',
      boxShadow: '0 0 20px rgba(88, 58, 255, 0.25)',
    },
    qrCodeContainer: {
      textAlign: 'center',
      padding: '1.5rem',
      background: 'rgba(21, 26, 38, 0.6)',
      border: '1px solid rgba(88, 58, 255, 0.2)',
      borderRadius: '1rem',
      marginBottom: '1.5rem',
    },
    qrCode: {
      width: '180px',
      height: '180px',
      margin: '0 auto 1rem',
    },
    copyButton: {
      background: 'rgba(21, 26, 38, 0.8)',
      border: '1px solid rgba(88, 58, 255, 0.2)',
      borderRadius: '0.75rem',
      padding: '0.75rem 1rem',
      color: '#F8F9FA',
      fontSize: '0.9rem',
      fontWeight: 500,
      cursor: 'pointer',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '0.5rem',
      transition: 'all 0.3s ease',
      width: '100%',
    },
    backButton: {
      background: 'none',
      border: 'none',
      color: '#B8BDC7',
      fontSize: '0.9rem',
      fontWeight: 600,
      cursor: 'pointer',
      display: 'flex',
      alignItems: 'center',
      gap: '0.5rem',
      marginBottom: '1rem',
      transition: 'all 0.3s ease',
    },
    amountPreview: {
      fontSize: 'var(--title-h4)',
      fontWeight: 800,
      background: 'linear-gradient(135deg, #F8F9FA 0%, #583AFF 50%, #1AD2FF 100%)',
      WebkitBackgroundClip: 'text',
      WebkitTextFillColor: 'transparent',
      textAlign: 'center',
      marginBottom: '1rem',
    },
  };

  if (!isOpen) return null;

  const handleNext = () => {
    if (currentStep === 1) {
      if (!depositAmount || parseFloat(depositAmount) <= 0) {
        alert('Por favor, informe um valor válido para depósito.');
        return;
      }
      setCurrentStep(2);
    } else if (currentStep === 2) {
      setCurrentStep(3);
    }
  };

  const handleConfirmDeposit = () => {
    if (onDepositSuccess) {
      onDepositSuccess(parseFloat(depositAmount));
    }
  };

  const handleCopyCode = async () => {
    const success = await copyToClipboard(fakePixCode);
    if (success) {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && currentStep < 3) {
      handleNext();
    }
  };

  return (
    <motion.div
      style={styles.overlay}
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      onClick={onClose}
    >
      <motion.div
        ref={modalRef}
        style={styles.modal}
        initial={{ scale: 0.8, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        exit={{ scale: 0.8, opacity: 0 }}
        onClick={(e) => e.stopPropagation()}
      >
        <button
          style={styles.closeButton}
          onClick={onClose}
          whileHover={{
            background: 'rgba(88, 58, 255, 0.2)',
            color: '#F8F9FA',
            boxShadow: '0 0 20px rgba(88, 58, 255, 0.4)',
          }}
          aria-label="Fechar">
          <X size={20} />
        </button>

        <div style={styles.header}>
          <h2 style={styles.title}>Adicionar Fundos</h2>
          <p style={styles.subtitle}>Escolha seu método de depósito</p>
        </div>

        {/* STEP INDICATOR */}
        <div style={styles.stepIndicator}>
          <div style={currentStep >= 1 ? { ...styles.stepDot, ...styles.stepDotActive } : styles.stepDot}></div>
          <div style={currentStep >= 2 ? { ...styles.stepDot, ...styles.stepDotActive } : styles.stepDot}></div>
          <div style={currentStep >= 3 ? { ...styles.stepDot, ...styles.stepDotActive } : styles.stepDot}></div>
        </div>

        {/* STEP 1: ENTER AMOUNT */}
        {currentStep === 1 && (
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -20 }}
          >
            <label style={{ display: 'block', color: '#B8BDC7', marginBottom: '0.5rem', fontSize: '0.9rem' }}>
              Valor do Depósito (R$)
            </label>
            <input
              type="number"
              value={depositAmount}
              onChange={(e) => setDepositAmount(e.target.value)}
              placeholder="0,00"
              style={styles.inputField}
              onKeyDown={handleKeyPress}
              autoFocus
            />

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '0.5rem', marginBottom: '2rem' }}>
              {[50, 100, 200].map(amount => (
                <motion.button
                  key={amount}
                  style={{
                    ...styles.inputField,
                    padding: '0.75rem',
                    fontSize: '1rem',
                    textAlign: 'center',
                    background: depositAmount == amount ? 'rgba(88, 58, 255, 0.2)' : 'rgba(21, 26, 38, 0.6)',
                    border: depositAmount == amount ? '1px solid rgba(88, 58, 255, 0.4)' : '1px solid rgba(88, 58, 255, 0.1)',
                  }}
                  whileHover={{
                    background: 'rgba(88, 58, 255, 0.2)',
                    borderColor: 'rgba(88, 58, 255, 0.3)',
                  }}
                  onClick={() => setDepositAmount(amount.toString())}
                >
                  R$ {amount}
                </motion.button>
              ))}
            </div>

            <motion.button
              style={styles.buttonPrimary}
              whileHover={{
                transform: 'translateY(-2px)',
                boxShadow: '0 12px 40px rgba(88, 58, 255, 0.6)',
              }}
              whileTap={{ scale: 0.98 }}
              onClick={handleNext}
            >
              Continuar
              <ArrowLeft size={20} style={{ transform: 'rotate(180deg)' }} />
            </motion.button>
          </motion.div>
        )}

        {/* STEP 2: SELECT PAYMENT METHOD */}
        {currentStep === 2 && (
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -20 }}
          >
            <motion.button
              style={styles.backButton}
              whileHover={{ color: '#F8F9FA' }}
              onClick={() => setCurrentStep(1)}
            >
              <ArrowLeft size={16} />
              Voltar
            </motion.button>

            <label style={{ display: 'block', color: '#B8BDC7', marginBottom: '1rem', fontSize: '0.9rem' }}>
              Selecione o método de pagamento
            </label>

            <motion.div
              style={
                selectedMethod === 'pix'
                  ? { ...styles.paymentMethod, ...styles.paymentMethodSelected }
                  : styles.paymentMethod
              }
              whileHover={{
                background: 'rgba(88, 58, 255, 0.1)',
                borderColor: 'rgba(88, 58, 255, 0.3)',
              }}
              onClick={() => setSelectedMethod('pix')}
            >
              <QrCode size={24} color="#1AD2FF" />
              <div style={{ textAlign: 'left' }}>
                <div style={{ color: '#F8F9FA', fontWeight: 600 }}>PIX</div>
                <div style={{ color: '#B8BDC7', fontSize: '0.8rem' }}>Transferência instantânea</div>
              </div>
            </motion.div>

            <motion.div
              style={
                selectedMethod === 'credit_card'
                  ? { ...styles.paymentMethod, ...styles.paymentMethodSelected }
                  : styles.paymentMethod
              }
              whileHover={{
                background: 'rgba(88, 58, 255, 0.1)',
                borderColor: 'rgba(88, 58, 255, 0.3)',
              }}
              onClick={() => setSelectedMethod('credit_card')}
            >
              <CreditCard size={24} color="#1AD2FF" />
              <div style={{ textAlign: 'left' }}>
                <div style={{ color: '#F8F9FA', fontWeight: 600 }}>Cartão de Crédito</div>
                <div style={{ color: '#B8BDC7', fontSize: '0.8rem' }}>Em até 12x com juros</div>
              </div>
            </motion.div>

            <motion.button
              style={styles.buttonPrimary}
              whileHover={{
                transform: 'translateY(-2px)',
                boxShadow: '0 12px 40px rgba(88, 58, 255, 0.6)',
              }}
              whileTap={{ scale: 0.98 }}
              onClick={handleNext}
            >
              Continuar
              <ArrowLeft size={20} style={{ transform: 'rotate(180deg)' }} />
            </motion.button>
          </motion.div>
        )}

        {/* STEP 3: QR CODE */}
        {currentStep === 3 && (
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -20 }}
          >
            <motion.button
              style={styles.backButton}
              whileHover={{ color: '#F8F9FA' }}
              onClick={() => setCurrentStep(2)}
            >
              <ArrowLeft size={16} />
              Voltar
            </motion.button>

            <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
              <div style={styles.amountPreview}>R$ {parseFloat(depositAmount).toFixed(2)}</div>
              <p style={{ color: '#B8BDC7' }}>Escaneie o QR code abaixo para efetuar o pagamento</p>
            </div>

            <div style={styles.qrCodeContainer}>
              {/* Placeholder para QR Code */}
              <div style={{
                ...styles.qrCode,
                background: 'rgba(255,255,255,0.9)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                borderRadius: '0.5rem',
                marginBottom: '1.5rem'
              }}>
                <QrCode size={100} color="#0A0E1A" />
              </div>

              <p style={{ color: '#B8BDC7', fontSize: '0.9rem', marginBottom: '1rem' }}>
                Código PIX copiado para transferência
              </p>

              <button
                style={styles.copyButton}
                onClick={handleCopyCode}
              >
                {copied ? <Check size={16} /> : <Copy size={16} />}
                {copied ? 'Copiado!' : 'Copiar código PIX'}
              </button>
            </div>

            <motion.button
              style={styles.buttonSecondary}
              whileHover={{
                transform: 'translateY(-2px)',
                boxShadow: '0 12px 40px rgba(88, 58, 255, 0.4)',
                background: 'rgba(88, 58, 255, 0.2)',
              }}
              whileTap={{ scale: 0.98 }}
              onClick={() => {
                handleConfirmDeposit();
              }}
            >
              Concluir Depósito
            </motion.button>
          </motion.div>
        )}
      </motion.div>
    </motion.div >
  );
};

export default DepositModal;
